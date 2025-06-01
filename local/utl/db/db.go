package db

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/logging"
	"time"

	"go.uber.org/dig"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Start(di *dig.Container) {
	db, err := gorm.Open(sqlite.Open("acc.db"), &gorm.Config{})
	if err != nil {
		logging.Panic("failed to connect database")
	}
	err = di.Provide(func() *gorm.DB {
		return db
	})
	if err != nil {
		logging.Panic("failed to bind database")
	}
	Migrate(db)
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&model.ApiModel{})
	if err != nil {
		logging.Panic("failed to migrate model.ApiModel")
	}
	err = db.AutoMigrate(&model.Server{})
	if err != nil {
		logging.Panic("failed to migrate model.Server")
	}
	err = db.AutoMigrate(&model.Config{})
	if err != nil {
		logging.Panic("failed to migrate model.Config")
	}
	err = db.AutoMigrate(&model.Track{})
	if err != nil {
		logging.Panic("failed to migrate model.Track")
	}
	err = db.AutoMigrate(&model.CarModel{})
	if err != nil {
		logging.Panic("failed to migrate model.CarModel")
	}
	err = db.AutoMigrate(&model.CupCategory{})
	if err != nil {
		logging.Panic("failed to migrate model.CupCategory")
	}
	err = db.AutoMigrate(&model.DriverCategory{})
	if err != nil {
		logging.Panic("failed to migrate model.DriverCategory")
	}
	err = db.AutoMigrate(&model.SessionType{})
	if err != nil {
		logging.Panic("failed to migrate model.SessionType")
	}
	err = db.AutoMigrate(&model.StateHistory{})
	if err != nil {
		logging.Panic("failed to migrate model.StateHistory")
	}
	err = db.AutoMigrate(&model.SteamCredentials{})
	if err != nil {
		logging.Panic("failed to migrate model.SteamCredentials")
	}
	err = db.AutoMigrate(&model.SystemConfig{})
	if err != nil {
		logging.Panic("failed to migrate model.SystemConfig")
	}
	db.FirstOrCreate(&model.ApiModel{Api: "Works"})

	Seed(db)
}

func Seed(db *gorm.DB) error {
	if err := seedTracks(db); err != nil {
		return err
	}
	if err := seedCarModels(db); err != nil {
		return err
	}
	if err := seedDriverCategories(db); err != nil {
		return err
	}
	if err := seedCupCategories(db); err != nil {
		return err
	}
	if err := seedSessionTypes(db); err != nil {
		return err
	}
	if err := seedServers(db); err != nil {
		return err
	}
	if err := seedSteamCredentials(db); err != nil {
		return err
	}
	if err := seedSystemConfigs(db); err != nil {
		return err
	}
	return nil
}

func seedSteamCredentials(db *gorm.DB) error {
	credentials := []model.SteamCredentials{
		{
			ID: 1,
			Username: "test",
			Password: "test",
			DateCreated: time.Now().UTC(),
		},
	}
	for _, credential := range credentials {
		if err := db.FirstOrCreate(&credential).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedServers(db *gorm.DB) error {
	servers := []model.Server{
		{ID: 1, Name: "ACC Server - Barcelona", ServiceName: "ACC-Barcelona", Path: "C:\\steamcmd\\acc", FromSteamCMD: true},
		{ID: 2, Name: "ACC Server - Monza", ServiceName: "ACC-Monza", Path: "C:\\steamcmd\\acc2", FromSteamCMD: true},
		{ID: 3, Name: "ACC Server - Spa", ServiceName: "ACC-Spa", Path: "C:\\steamcmd\\acc3", FromSteamCMD: true},
		{ID: 4, Name: "ACC Server - League", ServiceName: "ACC-League", Path: "C:\\steamcmd\\acc-league", FromSteamCMD: true},
	}

	for _, track := range servers {
		if err := db.FirstOrCreate(&track).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedTracks(db *gorm.DB) error {
	tracks := []model.Track{
		{Name: "monza", UniquePitBoxes: 29, PrivateServerSlots: 60},
		{Name: "zolder", UniquePitBoxes: 34, PrivateServerSlots: 50},
		{Name: "brands_hatch", UniquePitBoxes: 32, PrivateServerSlots: 50},
		{Name: "silverstone", UniquePitBoxes: 36, PrivateServerSlots: 60},
		{Name: "paul_ricard", UniquePitBoxes: 33, PrivateServerSlots: 80},
		{Name: "misano", UniquePitBoxes: 30, PrivateServerSlots: 50},
		{Name: "spa", UniquePitBoxes: 82, PrivateServerSlots: 82},
		{Name: "nurburgring", UniquePitBoxes: 30, PrivateServerSlots: 50},
		{Name: "barcelona", UniquePitBoxes: 29, PrivateServerSlots: 50},
		{Name: "hungaroring", UniquePitBoxes: 27, PrivateServerSlots: 50},
		{Name: "zandvoort", UniquePitBoxes: 25, PrivateServerSlots: 50},
		{Name: "kyalami", UniquePitBoxes: 40, PrivateServerSlots: 50},
		{Name: "mount_panorama", UniquePitBoxes: 36, PrivateServerSlots: 50},
		{Name: "suzuka", UniquePitBoxes: 51, PrivateServerSlots: 105},
		{Name: "laguna_seca", UniquePitBoxes: 30, PrivateServerSlots: 50},
		{Name: "imola", UniquePitBoxes: 30, PrivateServerSlots: 50},
		{Name: "oulton_park", UniquePitBoxes: 28, PrivateServerSlots: 50},
		{Name: "donington", UniquePitBoxes: 37, PrivateServerSlots: 50},
		{Name: "snetterton", UniquePitBoxes: 26, PrivateServerSlots: 50},
		{Name: "cota", UniquePitBoxes: 30, PrivateServerSlots: 70},
		{Name: "indianapolis", UniquePitBoxes: 30, PrivateServerSlots: 60},
		{Name: "watkins_glen", UniquePitBoxes: 30, PrivateServerSlots: 60},
		{Name: "valencia", UniquePitBoxes: 29, PrivateServerSlots: 50},
		{Name: "nurburgring_24h", UniquePitBoxes: 50, PrivateServerSlots: 110},
		{Name: "red_bull_ring", UniquePitBoxes: 50, PrivateServerSlots: 50},
	}

	for _, track := range tracks {
		if err := db.FirstOrCreate(&track).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedCarModels(db *gorm.DB) error {
	carModels := []model.CarModel{
		{Value: 0, CarModel: "Porsche 991 GT3 R"},
		{Value: 1, CarModel: "Mercedes-AMG GT3"},
		// ... Add all car models from your list
	}

	for _, cm := range carModels {
		if err := db.FirstOrCreate(&cm).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedDriverCategories(db *gorm.DB) error {
	categories := []model.DriverCategory{
		{Value: 3, Category: "Platinum"},
		{Value: 2, Category: "Gold"},
		{Value: 1, Category: "Silver"},
		{Value: 0, Category: "Bronze"},
	}

	for _, cat := range categories {
		if err := db.FirstOrCreate(&cat).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedCupCategories(db *gorm.DB) error {
	categories := []model.CupCategory{
		{Value: 0, Category: "Overall"},
		{Value: 1, Category: "ProAm"},
		{Value: 2, Category: "Am"},
		{Value: 3, Category: "Silver"},
		{Value: 4, Category: "National"},
	}

	for _, cat := range categories {
		if err := db.FirstOrCreate(&cat).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedSessionTypes(db *gorm.DB) error {
	sessionTypes := []model.SessionType{
		{Value: 0, SessionType: "Practice"},
		{Value: 4, SessionType: "Qualifying"},
		{Value: 10, SessionType: "Race"},
	}

	for _, st := range sessionTypes {
		if err := db.FirstOrCreate(&st).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedSystemConfigs(db *gorm.DB) error {
	configs := []model.SystemConfig{
		{
			Key:          model.ConfigKeySteamCMDPath,
			DefaultValue: "c:\\steamcmd\\steamcmd.exe",
			Description:  "Path to SteamCMD executable",
			DateModified: time.Now().UTC().Format(time.RFC3339),
		},
		{
			Key:          model.ConfigKeyNSSMPath,
			DefaultValue: ".\\nssm.exe",
			Description:  "Path to NSSM executable",
			DateModified: time.Now().UTC().Format(time.RFC3339),
		},
	}

	for _, config := range configs {
		var exists bool
		err := db.Model(&model.SystemConfig{}).
			Select("count(*) > 0").
			Where("key = ?", config.Key).
			Find(&exists).
			Error
		if err != nil {
			return err
		}

		if !exists {
			if err := db.Create(&config).Error; err != nil {
				return err
			}
			logging.Info("Seeded system config: %s", config.Key)
		}
	}

	return nil
}
