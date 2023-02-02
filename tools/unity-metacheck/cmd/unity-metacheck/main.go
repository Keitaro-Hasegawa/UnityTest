package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jessevdk/go-flags"
)

func isIgnorePath(path string) bool {
	if path[len(path)-1] == '~' || filepath.Ext(path) == ".tmp" {
		return true
	}

	dirPath := filepath.Dir(path)
	dirs := strings.Split(dirPath, string(os.PathSeparator))

	for _, dir := range dirs {
		if len(dir) == 0 {
			continue
		}
		dir = strings.ToLower(dir)
		if dir[0] == '.' || dir == "cvs" {
			return true
		}
	}

	return false
}

type assetFile struct {
	Path          string
	IsDir         bool
	IsEmpty       bool
	IsMeta        bool
	IsAssetBundle bool
}

type pathInfo struct {
	BasePath   string
	AssetFiles map[string]assetFile
}

func newPathInfo(basePath string) *pathInfo {
	return &pathInfo{
		BasePath:   basePath,
		AssetFiles: make(map[string]assetFile),
	}
}

func collectFilePath(rootPath, assetBundleDir, path string, excludeAssetBundleDirs []string, info os.FileInfo, pathInfo *pathInfo) {
	relpath, _ := filepath.Rel(rootPath, path)
	dirName := filepath.Dir(relpath)

	// mark parent directory as not emtpty
	if parentAssetFile, exists := pathInfo.AssetFiles[dirName]; exists {
		parentAssetFile.IsEmpty = false
		pathInfo.AssetFiles[dirName] = parentAssetFile
	}

	assetFile := assetFile{Path: relpath}

	// check if directory and initially regard as empty
	if info.IsDir() {
		assetFile.IsDir = true
		assetFile.IsEmpty = true
	} else {
		// check if .meta file
		if strings.ToLower(filepath.Ext(path)) == ".meta" {
			assetFile.IsMeta = true
		} else if filepath.HasPrefix(path, assetBundleDir) {
			exclude := false
			for _, excludeDir := range excludeAssetBundleDirs {
				if filepath.HasPrefix(path, excludeDir) {
					exclude = true
					break
				}
			}
			if !exclude {
				assetFile.IsAssetBundle = true
			}
		}
	}

	pathInfo.AssetFiles[relpath] = assetFile
}

func getAllAssetFilePaths(rootPath, assetBundleDir string, excludeAssetBundleDirs []string) (*pathInfo, error) {
	pathInfo := newPathInfo(rootPath)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if path == rootPath || isIgnorePath(path) {
			return nil
		}
		collectFilePath(rootPath, assetBundleDir, path, excludeAssetBundleDirs, info, pathInfo)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pathInfo, nil
}

const countOfMetaExt = len(".meta")

func validateFilePaths(info *pathInfo) result {
	namesOfAssetWithMeta := stringSet{}

	invalidMetaFiles := make([]string, 0)
	for name, assetFile := range info.AssetFiles {
		if assetFile.IsMeta {
			assetName := name[:len(name)-countOfMetaExt] // remove `.meta` extension
			if _, ok := info.AssetFiles[assetName]; ok {
				namesOfAssetWithMeta.Add(assetName)
			} else {
				// found .meta file without asset file
				invalidMetaFiles = append(invalidMetaFiles, name)
			}
		}
	}

	// check remaining asset files without .meta file
	invalidAssetFiles := make([]string, 0)
	for assetName, assetFile := range info.AssetFiles {
		if filepath.Base(assetName)[0] == '.' {
			continue
		}
		if !assetFile.IsMeta && !namesOfAssetWithMeta.Contains(assetName) {
			invalidAssetFiles = append(invalidAssetFiles, assetName)
		}
	}

	emptyFolders := make([]string, 0)
	for assetPath, assetFile := range info.AssetFiles {
		if assetFile.IsDir && assetFile.IsEmpty {
			emptyFolders = append(emptyFolders, assetPath)
		}
	}

	duplicateAssetBundleFiles := make(map[string][]string)
	for assetPath, assetFile := range info.AssetFiles {
		if assetFile.IsAssetBundle {
			name := filepath.Base(assetPath)
			if len(name) == 0 || name[0] == '.' {
				continue
			}
			addToMapList(duplicateAssetBundleFiles, name, assetPath)
		}
	}
	for name, assetList := range duplicateAssetBundleFiles {
		if len(assetList) <= 1 {
			delete(duplicateAssetBundleFiles, name)
		}
	}

	sort.Strings(invalidAssetFiles)
	sort.Strings(invalidMetaFiles)
	sort.Strings(emptyFolders)

	return result{
		BasePath:                  info.BasePath,
		InvalidAssetFiles:         invalidAssetFiles,
		InvalidMetaFiles:          invalidMetaFiles,
		EmptyFolders:              emptyFolders,
		DuplicateAssetBundleFiles: duplicateAssetBundleFiles,
	}
}

func addToMapList(mapList map[string][]string, key, value string) {
	list, exists := mapList[key]
	if !exists {
		list = make([]string, 0)
	}
	list = append(list, value)
	mapList[key] = list
}

func renderText(result result) {
	if !result.HasContent() {
		fmt.Println("[OK] No errors!")
		return
	}

	if result.HasInvalidAssetFiles() {
		fmt.Println("## metaファイルのないファイルやフォルダが存在します")
		for _, file := range result.InvalidAssetFiles {
			fmt.Println("- " + file)
		}
	}

	if result.HasInvalidMetaFiles() {
		fmt.Println("## ファイルやフォルダのないmetaファイルが存在します")
		for _, file := range result.InvalidMetaFiles {
			fmt.Println("- " + file)
		}
	}

	if result.HasEmptyFolders() {
		fmt.Println("## 空のフォルダが存在します")
		for _, file := range result.EmptyFolders {
			fmt.Println("- " + file)
		}
	}

	if result.HasDuplicateAssetBundleFiles() {
		fmt.Println("## AssetBundleに同じ名前のファイルが存在します")
		for fileName, fileList := range result.DuplicateAssetBundleFiles {
			if len(fileList) > 1 {
				fmt.Println("- " + fileName)
				for _, filePath := range fileList {
					fmt.Println("  - " + filePath)
				}
			}
		}
	}
}

func main() {
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		panic(err)
	}

	basePath, err := filepath.Abs(opts.TargetPath)
	if err != nil {
		panic(err)
	}

	assetBundleDir := filepath.Join(basePath, opts.AssetBundleDir)

	excludeAssetBundleDirs := make([]string, 0)
	for _, excludePath := range opts.ExcludeAssetBundleDirs {
		excludeAssetBundleDirs = append(excludeAssetBundleDirs, filepath.Join(basePath, excludePath))
	}

	info, err := getAllAssetFilePaths(basePath, assetBundleDir, excludeAssetBundleDirs)
	if err != nil {
		panic(err)
	}

	result := validateFilePaths(info)
	renderText(result)

	if result.HasContent() {
		os.Exit(1)
	}
}
