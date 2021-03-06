# aquago
A Digital Aquarium written in go - A toy project

Which is something like this (depending on your fishes and decorative elements!):

![aquago](doc/aquago.png)

## Aquarium Content
Our digital aquarium is consisted of four separate entities:

* The background - a static image which is the seabed
* The bubbles - some decorative bubbles as real aquariums have
* The BG items - decorative items which can be dragged and placed round the seabed
* The FG items - the "alive" organisms of the digital aquarium which "freely" move around

In order to run this application read the following steps.

## Prerequisites
* a supported Go version with modules enabled
* a Google Drive's account
* an installed C compiler (for [ebiten](https://ebiten.org/documents/install.html))

## Clone this repo
Clone this repo by running

```console
git clone https://github.com/mzampetakis/aquago.git
```

## Configuration

### Env Vars 📄
Before running the app all configuration must be set. The configuration can be added an OS's Env Vars or inside the provided `.env` file. The necessary variables are:

```
# GDrive Folders
# Folder ID to retrieve BG items
BGFOLDERID=1EDCChSY0ujjrpL-fK0S4R89zpWxLYcHS
# Folder ID to retrieve FG items
FGFOLDERID=1065nob35R6cIgB_aUIcW4fOwMV6D9oiS

#Saved Assets
# Image to use as background
BGIMAGE=assets/bg4.png
# Folder to save downloaded BG items
GDRIVEBGSFOLDER=assets/gdrive/bg/
# Folder to save downloaded FG items
GDRIVEFGSFOLDER=assets/gdrive/fg/
# Folder to save processed BG items
BGSFOLDER=assets/bg/
# Folder to save processed FG items
FGSFOLDER=assets/fg/
# Wether to remove saved assets - used to retrieve only the latest active assets
CLEARASSETS=true

#Google Drive Credentials file
GDRIVE_CREDS_FILE=credentials.json
```

### Google Drive oAuth 🗂
In order to setup the Google's Drive API oAuth please read the documentation provided in the [corresponding repo](https://github.com/mzampetakis/gogle-drive).

## Running the app 🏃‍♂️
When configuration is ready you can run the app by executing:

```console
go run main.go
```

After proceeding with the Google's Drive oAuth steps as [presented here](https://github.com/mzampetakis/gogle-drive) the application will start to download and transform the assets. While donwloading you will see an empty aquarium and your assets will gradually pop in.

## Results 🐠
If no items are found within your Google Drive (or while they are getting downloaded) you will see something like this:

![Empty Aquarium](doc/bubbles.gif)

If you have added assets in your Google Drive you will see something like this after all of them have downloaded (depending on your assets): 

![Full Aquarium](doc/aquago.gif)


## Gifts 🎁
All images used in this project have their background removed!!
This was not done by hand of course. Each image used in this app (either as BG or FG) after being retrieved from Google Drive is transformed! This transformation removes image's background (if possible - using a sample from the top left portion of the image and a threshold), crop's it to leave only the effective image and resizes to a valid size.

This transformation is presented in the following image:

![Image Transformation](doc/imgtransform.png)

Also, at runtime, application polls the Google's Drive folders for newly added images and if a new image is found, it is downloaded, transformed and then it just pops into the aquarium! That's quite convenient!

## Future Work 🐡
* Run without the need of Google Drive usage. Work with local images placed in the corresponding folders.
* Improve the background removal algorithm (needs better efficiency with decent performance)
* Add some background sound - as a real aquarium.

# Enjoy
I hope you enjoy this app at least as I did while developing and using it...

### Contributions are more than welcome in form of PRs or issues! But most importantly I would like to know if you enjoyed this!
