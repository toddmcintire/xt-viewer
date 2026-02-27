package main

import (
	"fmt"
	"strings"
	"os"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/toddmcintire/x4-viewer.git/x4"
)

func main() {

	var filePaths []string
	//var texture rl.Texture2D
	var textures []rl.Texture2D
	var pageIndex int
	var pageLimit int
	type ModeFlag int
	const (
		XTG ModeFlag = iota + 1 //1
		XTH //2
		XTC //3
		XTCH //4
	)

	rl.InitWindow(480, 800, "x4 viewer")
	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		//Update
		if rl.IsFileDropped() {
			droppedFiles := rl.LoadDroppedFiles()
			for _, v := range droppedFiles {
				filePaths = append(filePaths, v)				
			}

			if len(filePaths) > 0 {
				
				if strings.Contains(filePaths[0], ".xtg") {
					buf := make([]byte, 48000)
					x4.GetXTGData(filePaths[0], buf)
					expanded := x4.ExpandBitmap(buf)
					img := rl.NewImage(expanded, 480, 800, 1, rl.UncompressedGrayscale)
					textures = append(textures, rl.LoadTextureFromImage(img))
				}

				if strings.Contains(filePaths[0], ".xtc") || strings.Contains(filePaths[0], ".xtch") {
					//get file pointer
					filePT, openErr := os.Open(filePaths[0])
					if openErr != nil {
						panic("error opening file")
					}

					//get header for index offset and metadata offset
					header, headerErr := x4.GetXTCHeader(filePT)
					if headerErr != nil {
						fmt.Errorf("could not get header: %v", headerErr)
					}
					// //get metadata for chapter offset
					// metadata, metadataErr := x4.GetXTCMetadata(filePT, header.MetadataOffset)
					// if metadataErr != nil {
					// 	fmt.Errorf("could not get metadata: %v", metadataErr)
					// }
					//get array of pages
					pages, pagesErr := x4.GetXTCPage(filePT, header.IndexOffset, header.PageCount)
					if pagesErr != nil {
						fmt.Errorf("could not get pages: %v", pagesErr)
					}
					pageLimit = int(header.PageCount)
					//get picture array from pages
					pictures, pictureErr := x4.GetXTCPages(pages, filePT)
					if pictureErr != nil {
						fmt.Errorf("could not get pictures: %v", pictureErr)
					}
					//loop through pictures
					if xtgPicture, ok := pictures.([]x4.XTG); ok {
						for _, picture := range xtgPicture {
							//expand bits
							expanded := x4.ExpandBitmap(picture.Data[:])
							//load image from bits
							img := rl.NewImage(expanded, 480, 800, 1, rl.UncompressedGrayscale)
							//load texture from image and add texture to texture array
							textures = append(textures, rl.LoadTextureFromImage(img))
						}
					} else if xthPicture, ok := pictures.([]x4.XTH); ok {
						for _, picture := range xthPicture {
							//expand bits
							expanded := x4.ExpandXTHBitmap(picture.Data[:])
							//load image from bits
							img := rl.NewImage(expanded, 480, 800, 1, rl.UncompressedGrayscale)
							//load texture from image and add texture to texture array
							textures = append(textures, rl.LoadTextureFromImage(img))
						}
					} else {
						fmt.Printf("%T", pictures)
					}
					
					

				}

				
			}

			rl.UnloadDroppedFiles()
		}

		if (rl.IsKeyPressed(rl.KeyLeft)) {
			if pageIndex != 0 {
				pageIndex--
			}
		}
		if (rl.IsKeyPressed(rl.KeyRight)) {
			if (pageIndex != pageLimit-1) {
				pageIndex++
			}
		}

		//Draw
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		if len(filePaths) == 0 {
			rl.DrawText("Drop file", 200, 400, 20, rl.DarkGray)
		} else {
			// if (texture != rl.Texture2D{}) {
			// 	rl.DrawTexture(texture, 0, 0, rl.RayWhite)
			// } else if (len(textures) != 0) {
			// 	rl.DrawTexture(textures[0], 0, 0, rl.RayWhite)
			// }
			rl.DrawTexture(textures[pageIndex], 0,0,rl.RayWhite)
		}
		rl.EndDrawing()
	}
	//rl.UnloadTexture(texture)
	for _, unload := range textures {
		rl.UnloadTexture(unload)
	}
}