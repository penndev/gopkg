package maker

import (
	"slices"
	"strings"
	"testing"
)

func TestDictResolve(t *testing.T) {
	tests := []struct {
		in       string
		wantPath []string
		wantISP  string
	}{
		{"中国-广东-深圳 电信", []string{"中国", "广东", "深圳"}, "电信"},
		{"日本-大阪府-大阪市", []string{"日本", "大阪府", "大阪市"}, ""},
		{"中国-香港 华为云", []string{"香港"}, "华为云"},
		{"中国-澳门 Apple", []string{"澳门"}, "Apple"},
		{"中国-台湾-台北市 中华电信", []string{"台湾", "台北市"}, "中华电信"},
		{"中国-台湾苗栗市 中华电信/HiNet", []string{"台湾", "苗栗市"}, "中华电信"},
		{"中国-台湾台中县", []string{"台湾", "台中县"}, ""},
		{"中国-香港九龙油尖旺区佐敦道", []string{"香港", "九龙油尖旺区佐敦道"}, ""},
		{"中国-香港葵涌区 BGP多线数据中心", []string{"香港", "葵涌区"}, "BGP多线数据中心"},
		{"中国-香港中西区 VMSHELL", []string{"香港", "中西区"}, "VMSHELL"},
		{
			"中国-香港/中国-湖北-武汉/日本-东京都/日本-大阪府-大阪市 AGATHA/Anycast",
			[]string{"香港"},
			"AGATHA",
		},
		{
			"中国-上海-上海-杨浦区 中国电信/上海理工大学",
			[]string{"中国", "上海", "上海", "杨浦区"},
			"中国电信",
		},
		{"中国-江苏-南京 教育网/南京艺术学院", []string{"中国", "江苏", "南京"}, "教育网"},
		{"澳大利亚 APNIC/CloudFlare公共DNS服务器", []string{"澳大利亚"}, "APNIC"},
		{"中国-陕西-西安 电信/IDC机房", []string{"中国", "陕西", "西安"}, "电信"},
		{"IANA 保留地址", []string{"IANA"}, ""},
		{"未知 IANA", []string{"IANA"}, ""},
		{"保留地址 IETF", []string{"IANA"}, ""},
		{"未分配地址 APNIC", []string{"IANA"}, ""},
		{"未分配地址 ARIN", []string{"IANA"}, ""},
		{"CoreLink骨干网", []string{"IANA"}, ""},
		{"美国 CoreLink骨干网", []string{"美国"}, "CoreLink骨干网"},
		{"xTom", []string{"IANA"}, ""},
		{"美国-加利福尼亚州-圣克拉拉-圣荷西 xTom", []string{"美国", "加利福尼亚州", "圣克拉拉", "圣荷西"}, "xTom"},
		{"纯真网络 2026年07月29日IP数据", []string{"纯真网络"}, "2026年07月29日IP数据"},
	}
	for _, tt := range tests {
		d := newDict()
		areaID, ispID := d.resolve(tt.in)
		gotPath := areaPath(d, areaID)
		if !slices.Equal(gotPath, tt.wantPath) {
			t.Errorf("%q path=%v want %v", tt.in, gotPath, tt.wantPath)
		}
		gotISP := ispName(d, ispID)
		if gotISP != tt.wantISP {
			t.Errorf("%q isp=%q want %q", tt.in, gotISP, tt.wantISP)
		}
	}
}

func TestDictResolveHMTNotUnderChina(t *testing.T) {
	d := newDict()
	hk, _ := d.resolve("中国-香港 华为云")
	tw, _ := d.resolve("中国-台湾-台北市 中华电信")
	gd, _ := d.resolve("中国-广东 电信")
	mo, _ := d.resolve("中国-澳门")

	if parentName(d, hk) != "" {
		t.Fatalf("香港 parent=%q, want top-level", parentName(d, hk))
	}
	if parentName(d, tw) != "台湾" {
		t.Fatalf("台北市 parent=%q, want 台湾", parentName(d, tw))
	}
	if parentName(d, mo) != "" {
		t.Fatalf("澳门 parent=%q, want top-level", parentName(d, mo))
	}
	if got := areaPath(d, gd); !slices.Equal(got, []string{"中国", "广东"}) {
		t.Fatalf("广东 path=%v", got)
	}

	var chinaID uint32
	for _, a := range d.areas {
		if a.Name == "中国" && a.ParentID == 0 {
			chinaID = a.ID
			break
		}
	}
	if chinaID == 0 {
		t.Fatal("missing top-level 中国")
	}
	for _, a := range d.areas {
		if a.ParentID == chinaID && isHMT(a.Name) {
			t.Fatalf("%s still under 中国", a.Name)
		}
	}

	hk2, _ := d.resolve("香港")
	if hk2 != hk {
		t.Fatalf("香港 vs 中国-香港: %d != %d", hk2, hk)
	}

	stuck, _ := d.resolve("中国-台湾苗栗市")
	if parentName(d, stuck) != "台湾" {
		t.Fatalf("台湾苗栗市 parent=%q, want 台湾", parentName(d, stuck))
	}
}

func TestDictPeelConcat(t *testing.T) {
	d := newDict()
	d.learn("中国-云南-红河哈尼族彝族自治州-蒙自市 联通")
	d.learn("中国-云南-大理白族自治州-漾濞彝族自治县 电信")
	d.learn("中国-河北-石家庄-辛集市 移动")
	d.learn("中国-吉林-长春 电信")
	d.learn("中国-吉林省 联通")
	d.learn("中国-云南红河哈尼族彝族自治州蒙自市 电信")
	d.learn("中国-河北石家庄辛集市 移动")
	d.dropConcatChildren()

	id, isp := d.resolve("中国-云南红河哈尼族彝族自治州蒙自市 电信")
	got := areaPath(d, id)
	want := []string{"中国", "云南", "红河哈尼族彝族自治州", "蒙自市"}
	if !slices.Equal(got, want) {
		t.Fatalf("蒙自 path=%v want %v", got, want)
	}
	if ispName(d, isp) != "电信" {
		t.Fatalf("isp=%q", ispName(d, isp))
	}

	id, _ = d.resolve("中国-云南大理白族自治州漾濞彝族自治县 电信")
	got = areaPath(d, id)
	want = []string{"中国", "云南", "大理白族自治州", "漾濞彝族自治县"}
	if !slices.Equal(got, want) {
		t.Fatalf("漾濞 path=%v want %v", got, want)
	}

	id, _ = d.resolve("中国-河北石家庄辛集市 移动")
	got = areaPath(d, id)
	want = []string{"中国", "河北", "石家庄", "辛集市"}
	if !slices.Equal(got, want) {
		t.Fatalf("辛集 path=%v want %v", got, want)
	}

	id, _ = d.resolve("中国-吉林省 联通")
	got = areaPath(d, id)
	if !slices.Equal(got, []string{"中国", "吉林省"}) {
		t.Fatalf("吉林省 should stay whole, got %v", got)
	}

	for _, a := range d.areas {
		if strings.Contains(a.Name, "云南红河") || strings.Contains(a.Name, "河北石家庄辛") {
			t.Fatalf("concat leftover %q", a.Name)
		}
	}
}

func areaPath(d *dict, id uint32) []string {
	var names []string
	for id != 0 {
		a := d.areas[id-1]
		names = append([]string{a.Name}, names...)
		id = a.ParentID
	}
	return names
}

func parentName(d *dict, id uint32) string {
	if id == 0 {
		return ""
	}
	pid := d.areas[id-1].ParentID
	if pid == 0 {
		return ""
	}
	return d.areas[pid-1].Name
}

func ispName(d *dict, id uint32) string {
	if id == 0 {
		return ""
	}
	return d.isps[id-1].Name
}
