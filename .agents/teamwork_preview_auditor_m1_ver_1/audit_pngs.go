package main

import (
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 1. Inspect context/
	contextPNGs := make(map[string]string) // relPath -> sha256
	err := filepath.WalkDir("context", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".png") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel("context", path)
			normRel := strings.ReplaceAll(rel, "90┬║ Rotatable Bridge Sprites", "90 Rotatable Bridge Sprites")
			contextPNGs[normRel] = fmt.Sprintf("%x", sha256.Sum256(data))
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking context: %v\n", err)
	}

	fmt.Printf("Context PNG count: %d\n", len(contextPNGs))

	// 2. Inspect internal/assets/images/
	imagesPNGs := make(map[string]string) // relPath -> sha256
	var legacyCount, externalCount int
	var decodeFailures []string
	var zeroSize []string
	var nonPngHeaders []string

	pngHeader := []byte("\x89PNG\r\n\x1a\n")

	err = filepath.WalkDir("internal/assets/images", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".png") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				zeroSize = append(zeroSize, path)
			}
			if len(data) < 8 || string(data[:8]) != string(pngHeader) {
				nonPngHeaders = append(nonPngHeaders, path)
			}
			rel, _ := filepath.Rel("internal/assets/images", path)
			hash := fmt.Sprintf("%x", sha256.Sum256(data))
			imagesPNGs[rel] = hash

			// Try image.Decode
			f, err := os.Open(path)
			if err != nil {
				decodeFailures = append(decodeFailures, fmt.Sprintf("%s open err: %v", path, err))
			} else {
				img, format, err := image.Decode(f)
				f.Close()
				if err != nil {
					decodeFailures = append(decodeFailures, fmt.Sprintf("%s decode err: %v", path, err))
				} else if format != "png" {
					decodeFailures = append(decodeFailures, fmt.Sprintf("%s unexpected format: %s", path, format))
				} else {
					b := img.Bounds()
					if b.Dx() <= 0 || b.Dy() <= 0 {
						decodeFailures = append(decodeFailures, fmt.Sprintf("%s invalid dimensions: %dx%d", path, b.Dx(), b.Dy()))
					}
				}
			}

			// Check if legacy vs external
			if strings.HasPrefix(rel, "Lab/") || strings.HasPrefix(rel, "Small Forest/") || strings.HasPrefix(rel, "Zombie Apocalypse Tileset/") {
				externalCount++
			} else {
				legacyCount++
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking images: %v\n", err)
	}

	fmt.Printf("Total images/ PNG count: %d (Legacy: %d, External: %d)\n", len(imagesPNGs), legacyCount, externalCount)
	fmt.Printf("Non-PNG headers count: %d\n", len(nonPngHeaders))
	fmt.Printf("Decode failures count: %d\n", len(decodeFailures))
	if len(decodeFailures) > 0 {
		for _, f := range decodeFailures {
			fmt.Printf("  Fail: %s\n", f)
		}
	}
	fmt.Printf("Zero size files: %d\n", len(zeroSize))

	// 3. Match external PNGs with context PNGs
	mismatchCount := 0
	missingInImages := 0
	for ctxRel, ctxHash := range contextPNGs {
		imgHash, exists := imagesPNGs[ctxRel]
		if !exists {
			fmt.Printf("Missing in images: %s\n", ctxRel)
			missingInImages++
		} else if imgHash != ctxHash {
			fmt.Printf("Hash mismatch for %s: context=%s, images=%s\n", ctxRel, ctxHash, imgHash)
			mismatchCount++
		}
	}
	fmt.Printf("Summary vs context: missing=%d, hash_mismatch=%d\n", missingInImages, mismatchCount)

	// 4. Check if there are any unexpected external files in internal/assets/images that are not in context
	extraExternal := 0
	for imgRel := range imagesPNGs {
		if strings.HasPrefix(imgRel, "Lab/") || strings.HasPrefix(imgRel, "Small Forest/") || strings.HasPrefix(imgRel, "Zombie Apocalypse Tileset/") {
			if _, exists := contextPNGs[imgRel]; !exists {
				fmt.Printf("Extra external file in images: %s\n", imgRel)
				extraExternal++
			}
		}
	}
	fmt.Printf("Extra external files: %d\n", extraExternal)
}
