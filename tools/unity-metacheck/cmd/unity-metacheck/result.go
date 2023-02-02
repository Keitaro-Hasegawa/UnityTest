package main

type result struct {
	BasePath                  string              `json:"base_path"`
	InvalidAssetFiles         []string            `json:"invalid_asset_files"`
	InvalidMetaFiles          []string            `json:"invalid_meta_files"`
	EmptyFolders              []string            `json:"empty_folders"`
	DuplicateAssetBundleFiles map[string][]string `json:"duplicate_assetbundle_files"`
}

func (r result) HasContent() bool {
	return r.HasInvalidAssetFiles() ||
		r.HasInvalidMetaFiles() ||
		r.HasEmptyFolders() ||
		r.HasDuplicateAssetBundleFiles()
}

func (r result) HasInvalidAssetFiles() bool {
	return len(r.InvalidAssetFiles) > 0
}

func (r result) HasInvalidMetaFiles() bool {
	return len(r.InvalidMetaFiles) > 0
}

func (r result) HasEmptyFolders() bool {
	return len(r.EmptyFolders) > 0
}

func (r result) HasDuplicateAssetBundleFiles() bool {
	for _, list := range r.DuplicateAssetBundleFiles {
		if len(list) > 1 {
			return true
		}
	}
	return false
}
