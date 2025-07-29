package file_handler

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

)

func BuildIpeFromFolder(version string,sourceDir string, outputDir string) error {
	manifest, err := CreateManifest(sourceDir, version)
	if err != nil{
		return err
	}
	outfile, err := os.Create(filepath.Join(outputDir,manifest.Tool.Name + ".ipe"))
	if err != nil{
		return err
	}
	defer outfile.Close()
	
	writer := zip.NewWriter(outfile)
	defer writer.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir(){
			return nil
		}
		relpath, err := filepath.Rel(sourceDir,path)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipEntryWriter, err := writer.Create(relpath)
		if err != nil {
			return err
		}
		_,err = io.Copy(zipEntryWriter,file)
		return err
	})
}