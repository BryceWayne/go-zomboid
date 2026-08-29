package assets

import (
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type assetDescriptor struct {
	name  string
	ptr   **ebiten.Image
	path  string
	wantW int
	wantH int
	cat   string
}

func getAssetDescriptors() []assetDescriptor {
	return []assetDescriptor{
		// Character Entities (3) - 64x128
		{"PlayerImage", &PlayerImage, "images/player.png", 64, 128, "Entity"},
		{"ZombieImage", &ZombieImage, "images/zombie.png", 64, 128, "Entity"},
		{"RunnerImage", &RunnerImage, "images/runner.png", 64, 128, "Entity"},

		// Floor Tiles (6) - 256x128
		{"GrassImage", &GrassImage, "images/grass.png", 256, 128, "Floor"},
		{"DirtImage", &DirtImage, "images/dirt.png", 256, 128, "Floor"},
		{"WoodImage", &WoodImage, "images/wood.png", 256, 128, "Floor"},
		{"AsphaltImage", &AsphaltImage, "images/asphalt.png", 256, 128, "Floor"},
		{"ConcreteImage", &ConcreteImage, "images/concrete.png", 256, 128, "Floor"},
		{"TileFloorImage", &TileFloorImage, "images/tile_floor.png", 256, 128, "Floor"},

		// Vertical Obstacles / Props (10) - 256x256
		{"WallImage", &WallImage, "images/wall.png", 256, 256, "Obstacle/Prop"},
		{"TreeImage", &TreeImage, "images/tree.png", 256, 256, "Obstacle/Prop"},
		{"FenceImage", &FenceImage, "images/fence.png", 256, 256, "Obstacle/Prop"},
		{"DebrisImage", &DebrisImage, "images/debris.png", 256, 256, "Obstacle/Prop"},
		{"TentImage", &TentImage, "images/tent.png", 256, 256, "Obstacle/Prop"},
		{"StumpImage", &StumpImage, "images/stump.png", 256, 256, "Obstacle/Prop"},
		{"MushroomImage", &MushroomImage, "images/mushroom.png", 256, 256, "Obstacle/Prop"},
		{"SignImage", &SignImage, "images/sign.png", 256, 256, "Obstacle/Prop"},
		{"ElevationBlockImage", &ElevationBlockImage, "images/elevation_block.png", 256, 256, "Obstacle/Prop"},
		{"ElevationRampImage", &ElevationRampImage, "images/elevation_ramp.png", 256, 256, "Obstacle/Prop"},

		// Item / Weapon / Equipment (8) - 64x64
		{"FoodImage", &FoodImage, "images/food.png", 64, 64, "Item"},
		{"WaterImage", &WaterImage, "images/water.png", 64, 64, "Item"},
		{"WeaponImage", &WeaponImage, "images/weapon.png", 64, 64, "Item"},
		{"AxeImage", &AxeImage, "images/axe.png", 64, 64, "Item"},
		{"ShotgunImage", &ShotgunImage, "images/shotgun.png", 64, 64, "Item"},
		{"AmmoImage", &AmmoImage, "images/ammo.png", 64, 64, "Item"},
		{"ArmorImage", &ArmorImage, "images/armor.png", 64, 64, "Item"},
		{"AntidoteImage", &AntidoteImage, "images/antidote.png", 64, 64, "Item"},

		// External World Props & Foliage (from context/)
		{"BenchImage", &BenchImage, "images/Small Forest/Bench and chest/Bench.png", 52, 37, "ExternalProp"},
		{"ChestImage", &ChestImage, "images/Small Forest/Bench and chest/Chest.png", 22, 21, "ExternalProp"},
		{"Sculpture1Image", &Sculpture1Image, "images/Small Forest/Sculptures/Sculpture-1.png", 23, 31, "ExternalProp"},
		{"Sculpture2Image", &Sculpture2Image, "images/Small Forest/Sculptures/Sculture-2.png", 29, 32, "ExternalProp"},
		{"SculptureImage", &SculptureImage, "images/Small Forest/Sculptures/Sculpture-1.png", 23, 31, "ExternalProp"},
		{"Bush1Image", &Bush1Image, "images/Small Forest/Bushes/Bush-1.png", 24, 18, "ExternalProp"},
		{"Bush2Image", &Bush2Image, "images/Small Forest/Bushes/Bush-2.png", 19, 15, "ExternalProp"},
		{"Bush3Image", &Bush3Image, "images/Small Forest/Bushes/Bush-3.png", 25, 19, "ExternalProp"},
		{"Bush4Image", &Bush4Image, "images/Small Forest/Bushes/Bush-4.png", 28, 19, "ExternalProp"},
		{"BushImage", &BushImage, "images/Small Forest/Bushes/Bush-1.png", 24, 18, "ExternalProp"},
		{"Flower1Image", &Flower1Image, "images/Small Forest/Flowers/Flower-1.png", 26, 25, "ExternalProp"},
		{"Flower2Image", &Flower2Image, "images/Small Forest/Flowers/Flower-2.png", 24, 22, "ExternalProp"},
		{"Flower3Image", &Flower3Image, "images/Small Forest/Flowers/Flower-3.png", 26, 18, "ExternalProp"},
		{"FlowerImage", &FlowerImage, "images/Small Forest/Flowers/Flower-1.png", 26, 25, "ExternalProp"},
		{"Stone1Image", &Stone1Image, "images/Small Forest/Stones/Stone-1.png", 28, 19, "ExternalProp"},
		{"Stone2Image", &Stone2Image, "images/Small Forest/Stones/Stone-2.png", 29, 25, "ExternalProp"},
		{"StoneImage", &StoneImage, "images/Small Forest/Stones/Stone-1.png", 28, 19, "ExternalProp"},
		{"ForestStumpImage", &ForestStumpImage, "images/Small Forest/Bushes/Stump.png", 29, 19, "ExternalProp"},
		{"GrassTuft1Image", &GrassTuft1Image, "images/Small Forest/Grass/Grass-1.png", 25, 24, "ExternalProp"},
		{"GrassTuft2Image", &GrassTuft2Image, "images/Small Forest/Grass/Grass-2.png", 31, 15, "ExternalProp"},

		// External Tilesets
		{"LabTilesetImage", &LabTilesetImage, "images/Lab/Inside_C.png", 768, 768, "ExternalTileset"},
		{"ZombieTilesetImage", &ZombieTilesetImage, "images/Zombie Apocalypse Tileset/Zombie Apocalypse Tileset Reference.png", 764, 300, "ExternalTileset"},
	}
}

func TestChallenger_AllExportedPointersAndExactBounds(t *testing.T) {
	Load()

	descriptors := getAssetDescriptors()
	if len(descriptors) != 49 {
		t.Fatalf("expected exactly 49 exported image descriptors (27 legacy + 22 external), got %d", len(descriptors))
	}

	for _, d := range descriptors {
		t.Run(d.name, func(t *testing.T) {
			img := *d.ptr
			if img == nil {
				t.Fatalf("Pointer %s is nil after Load()", d.name)
			}

			bounds := img.Bounds()
			if bounds.Min.X != 0 || bounds.Min.Y != 0 {
				t.Errorf("Pointer %s Bounds().Min = (%d, %d), want (0, 0)", d.name, bounds.Min.X, bounds.Min.Y)
			}
			if bounds.Dx() != d.wantW || bounds.Dy() != d.wantH {
				t.Errorf("Pointer %s Bounds() = %dx%d, want %dx%d (%s category)",
					d.name, bounds.Dx(), bounds.Dy(), d.wantW, d.wantH, d.cat)
			}
		})
	}
}

func TestChallenger_MultiThreadedLoadAndPointerRace(t *testing.T) {
	const numLoaders = 20
	const numReaders = 30
	const iterations = 50

	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	for i := 0; i < numLoaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal
			for j := 0; j < iterations; j++ {
				Load()
			}
		}()
	}

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			<-startSignal
			for j := 0; j < iterations; j++ {
				descriptors := getAssetDescriptors()
				d := descriptors[readerID%len(descriptors)]
				img := *d.ptr
				if img != nil {
					b := img.Bounds()
					if b.Dx() != d.wantW || b.Dy() != d.wantH {
						t.Errorf("Reader %d detected bounds mismatch: %dx%d want %dx%d",
							readerID, b.Dx(), b.Dy(), d.wantW, d.wantH)
					}
				}
			}
		}(i)
	}

	close(startSignal)
	wg.Wait()
}
