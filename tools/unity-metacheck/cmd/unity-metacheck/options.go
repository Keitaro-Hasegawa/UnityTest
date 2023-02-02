package main

type options struct {
	TargetPath             string   `short:"p" long:"target-path" description:"Unity assets path" required:"yes"`
	AssetBundleDir         string   `short:"a" description:"AssetBundle directory"`
	ExcludeAssetBundleDirs []string `short:"e" description:"Exclude AssetBundle directories"`
	FormatJSON             bool     `short:"j" long:"json" description:"output json format"`
}
