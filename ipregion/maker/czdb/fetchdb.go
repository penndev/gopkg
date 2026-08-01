package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/penndev/gopkg/ipregion/maker"
	"github.com/penndev/gopkg/ipregion/maker/czdb/search/db"
)

func ensureCZDB(workDir, zipURL string) error {
	if maker.FileReady(filepath.Join(workDir, czdbV4)) && maker.FileReady(filepath.Join(workDir, czdbV6)) {
		log.Println("已存在 czdb 文件，跳过下载")
		return nil
	}
	if zipURL == "" {
		return fmt.Errorf("缺少 czdb 文件，且未提供 ZipURL（可用环境变量 CZDB_FILE）")
	}
	zipPath := filepath.Join(workDir, "czdb.zip")
	log.Printf("下载: %s", zipURL)
	if err := download(zipURL, zipPath); err != nil {
		return fmt.Errorf("下载: %w", err)
	}
	if err := extractCZDB(zipPath, workDir); err != nil {
		return fmt.Errorf("解压: %w", err)
	}
	return nil
}

func verifyCZDB(workDir, key string) error {
	tests := []struct {
		file string
		ip   string
	}{
		{czdbV4, "8.8.8.8"},
		{czdbV6, "2001:4860:4860::8888"},
	}
	for _, t := range tests {
		path := filepath.Join(workDir, t.file)
		s, err := db.InitDBSearcher(path, key, db.MEMORY)
		if err != nil {
			return fmt.Errorf("密钥无效或文件损坏 (%s): %w", t.file, err)
		}
		region, err := db.Search(t.ip, s)
		db.CloseDBSearcher(s)
		if err != nil {
			return fmt.Errorf("查询失败 (%s, %s): %w", t.file, t.ip, err)
		}
		log.Printf("通过: %s  %s -> %s", t.file, t.ip, region)
	}
	return nil
}

func download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s", resp.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func extractCZDB(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	need := map[string]string{
		czdbV4: filepath.Join(destDir, czdbV4),
		czdbV6: filepath.Join(destDir, czdbV6),
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		dest, ok := need[filepath.Base(f.Name)]
		if !ok {
			continue
		}
		if err := copyZipFile(f, dest); err != nil {
			return err
		}
		delete(need, filepath.Base(f.Name))
	}
	for name := range need {
		return fmt.Errorf("zip 中缺少 %s", name)
	}
	return nil
}

func copyZipFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}
