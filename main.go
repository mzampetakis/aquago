package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/joho/godotenv"
	"github.com/mzampetakis/aquago/aquarium"
	"github.com/mzampetakis/aquago/transimage"
	gogledrive "github.com/mzampetakis/gogle-drive"
)

func init() {
	loadEnvVars()
	if os.Getenv("GDRIVE_CREDS_FILE") == "" {
		log.Fatalf("Could not load GDRIVE_CREDS_FILE env var")
	}
	if os.Getenv("BGSFOLDER") == "" {
		log.Fatalf("Could not load BGSFOLDER env var")
	}
	if os.Getenv("FGSFOLDER") == "" {
		log.Fatalf("Could not load FGSFOLDER env var")
	}
	if os.Getenv("GDRIVEBGSFOLDER") == "" {
		log.Fatalf("Could not load GDRIVEBGSFOLDER env var")
	}
	if os.Getenv("GDRIVEFGSFOLDER") == "" {
		log.Fatalf("Could not load GDRIVEFGSFOLDER env var")
	}
}

func loadEnvVars() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println(".env file not found")
	}
}

func main() {
	fmt.Println("Welcome to AquaGo!")
	bgImage := os.Getenv("BGIMAGE")

	gdriveCredsFile := os.Getenv("GDRIVE_CREDS_FILE")
	bgFolder := os.Getenv("BGSFOLDER")
	fgFolder := os.Getenv("FGSFOLDER")
	gdriveBgFolder := os.Getenv("GDRIVEBGSFOLDER")
	gdriveFgFolder := os.Getenv("GDRIVEFGSFOLDER")

	clearAssets := os.Getenv("CLEARASSETS")
	if clearAssets == "true" {
		fmt.Println("Clearing assets...")
		foldersToEmpty := []string{bgFolder, fgFolder, gdriveBgFolder, gdriveFgFolder}
		for _, folderToEmpty := range foldersToEmpty {
			removeFilesFrom(folderToEmpty)
		}
	}

	fmt.Println("Initiating assets download...")
	fmt.Println("Your assets will appear in a while...")
	gogledrive, err := gogledrive.New(gdriveCredsFile)
	if err != nil {
		log.Fatalf("Could not instantiate google drive %s", err)
	}
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			downloadNewImages(gogledrive, os.Getenv("BGFOLDERID"), gdriveBgFolder, bgFolder)
			downloadNewImages(gogledrive, os.Getenv("FGFOLDERID"), gdriveFgFolder, fgFolder)
		}
	}()

	fmt.Println("Starting Aquarium. Enjoy!")
	aquarium.Start(bgImage, bgFolder, fgFolder)
}

func removeFilesFrom(dirPath string) {
	dir, _ := ioutil.ReadDir(dirPath)
	for _, d := range dir {
		os.RemoveAll(path.Join([]string{dirPath, d.Name()}...))
	}
}

func downloadNewImages(gdrive *gogledrive.Gogledrive, gdriveFolderID string, gdriveFolderPath string, folderPath string) {
	imageMimeType := "image/"
	imagesFiler := gogledrive.ListFilter{
		FolderID: &gdriveFolderID,
		MimeType: &imageMimeType,
	}
	gdriveImagesList, err := gdrive.SearchFiles(imagesFiler)
	if err != nil {
		log.Fatal(err)
	}
	for imageName, imageID := range gdriveImagesList {
		//if exists in gdriveFolderPath continue
		alreadyProcessed := false
		dir, _ := ioutil.ReadDir(gdriveFolderPath)
		for _, d := range dir {
			if d.Name() == imageName {
				alreadyProcessed = true
				break
			}
		}
		if alreadyProcessed {
			continue
		}
		imgBuf, err := gdrive.GetFile(imageID)
		if err != nil {
			fmt.Println(err)
		}
		err = transimage.SaveBytesToImageFile(imgBuf, gdriveFolderPath+imageName)
		if err == nil {
			transimage.RemoveBG(gdriveFolderPath+imageName, folderPath)
		}
	}
}
