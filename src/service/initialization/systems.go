package initialization

import (
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/repository"
	"aegis/service/common"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// builtinSystems defines the 6 built-in systems that are seeded on startup.
var builtinSystems = []database.System{
	{Name: "train-ticket", DisplayName: "Train Ticket", NsPattern: `^ts\d+$`, ExtractPattern: `^(ts)(\d+)$`, Count: 1, IsBuiltin: true, Status: consts.CommonEnabled},
	{Name: "sock-shop", DisplayName: "Sock Shop", NsPattern: `^ss\d+$`, ExtractPattern: `^(ss)(\d+)$`, Count: 1, IsBuiltin: true, Status: consts.CommonEnabled},
	{Name: "social-network", DisplayName: "Social Network", NsPattern: `^sn\d+$`, ExtractPattern: `^(sn)(\d+)$`, Count: 1, IsBuiltin: true, Status: consts.CommonEnabled},
	{Name: "online-boutique", DisplayName: "Online Boutique", NsPattern: `^ob\d+$`, ExtractPattern: `^(ob)(\d+)$`, Count: 1, IsBuiltin: true, Status: consts.CommonEnabled},
	{Name: "hotel-reservation", DisplayName: "Hotel Reservation", NsPattern: `^hr\d+$`, ExtractPattern: `^(hr)(\d+)$`, Count: 1, IsBuiltin: true, Status: consts.CommonEnabled},
	{Name: "media-microsvc", DisplayName: "Media Microservices", NsPattern: `^mm\d+$`, ExtractPattern: `^(mm)(\d+)$`, Count: 1, IsBuiltin: true, Status: consts.CommonEnabled},
}

// InitializeSystems seeds built-in systems, registers all enabled systems with
// chaos-experiment, and sets the global MetadataStore.
func InitializeSystems() {
	// Set DB reference for ChaosSystemConfig to query System table
	config.SetChaosConfigDB(database.DB)

	// Seed built-in systems using FirstOrCreate
	for _, sys := range builtinSystems {
		var existing database.System
		result := database.DB.Where("name = ?", sys.Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := database.DB.Create(&sys).Error; err != nil {
				logrus.Warnf("Failed to seed builtin system %s: %v", sys.Name, err)
			} else {
				logrus.Infof("Seeded builtin system: %s", sys.Name)
			}
		}
	}

	// Load all enabled systems from DB and register with chaos-experiment
	systems, err := repository.ListEnabledSystems(database.DB)
	if err != nil {
		logrus.Errorf("Failed to load enabled systems: %v", err)
		return
	}

	for _, sys := range systems {
		if err := chaos.RegisterSystem(chaos.SystemConfig{
			Name:        sys.Name,
			NsPattern:   sys.NsPattern,
			DisplayName: sys.DisplayName,
		}); err != nil {
			logrus.Warnf("Failed to register system %s: %v", sys.Name, err)
		} else {
			logrus.Infof("Registered system: %s (%s)", sys.Name, sys.DisplayName)
		}
	}

	// Create and set the global MetadataStore
	store := common.NewDBMetadataStore()
	chaos.SetMetadataStore(store)
	logrus.Info("Set global DBMetadataStore for chaos-experiment")
}
