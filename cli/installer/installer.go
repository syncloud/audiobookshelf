package installer

import (
	"fmt"
	cp "github.com/otiai10/copy"
	"github.com/syncloud/golib/config"
	"github.com/syncloud/golib/linux"
	"github.com/syncloud/golib/platform"
	"go.uber.org/zap"
	"os"
	"path"
)

type Variables struct {
	App          string
	AppDir       string
	DataDir      string
	CommonDir    string
	StorageDir   string
	Socket       string
}

const (
	App       = "audiobookshelf"
	AppDir    = "/snap/audiobookshelf/current"
	DataDir   = "/var/snap/audiobookshelf/current"
	CommonDir = "/var/snap/audiobookshelf/common"
)

type Installer struct {
	newVersionFile     string
	currentVersionFile string
	platformClient     *platform.Client
	installFile        string
	logger             *zap.Logger
	oidc               *Oidc
}

func New(logger *zap.Logger) *Installer {
	platformClient := platform.New()
	return &Installer{
		newVersionFile:     path.Join(AppDir, "version"),
		currentVersionFile: path.Join(DataDir, "version"),
		platformClient:     platformClient,
		installFile:        path.Join(CommonDir, "installed"),
		logger:             logger,
		oidc:               NewOidc(platformClient, logger, DataDir),
	}
}

func (i *Installer) Install() error {
	return i.UpdateConfigs()
}

func (i *Installer) Configure() error {
	if err := i.StorageChange(); err != nil {
		return err
	}
	if !i.IsInstalled() {
		if err := i.Initialize(); err != nil {
			return err
		}
	}
	return i.UpdateVersion()
}

func (i *Installer) IsInstalled() bool {
	_, err := os.Stat(i.installFile)
	return err == nil
}

func (i *Installer) Initialize() error {
	storageDir, err := i.platformClient.InitStorage(App, App)
	if err != nil {
		return err
	}
	if err := i.oidc.ConfigureApp(storageDir); err != nil {
		return fmt.Errorf("configure app: %w", err)
	}
	return os.WriteFile(i.installFile, []byte("installed"), 0644)
}

func (i *Installer) PreRefresh() error {
	return nil
}

func (i *Installer) PostRefresh() error {
	if err := i.UpdateConfigs(); err != nil {
		return err
	}
	if err := i.ClearVersion(); err != nil {
		return err
	}
	return i.FixPermissions()
}

func (i *Installer) AccessChange() error {
	return i.UpdateConfigs()
}

func (i *Installer) StorageChange() error {
	storageDir, err := i.platformClient.InitStorage(App, App)
	if err != nil {
		return err
	}
	if err := linux.CreateMissingDirs(
		path.Join(storageDir, "config"),
		path.Join(storageDir, "metadata"),
	); err != nil {
		return err
	}
	return linux.Chown(storageDir, App)
}

func (i *Installer) ClearVersion() error {
	return os.RemoveAll(i.currentVersionFile)
}

func (i *Installer) UpdateVersion() error {
	return cp.Copy(i.newVersionFile, i.currentVersionFile)
}

func (i *Installer) UpdateConfigs() error {
	if err := linux.CreateUser(App); err != nil {
		return err
	}
	if err := i.StorageChange(); err != nil {
		return err
	}
	if err := linux.CreateMissingDirs(path.Join(DataDir, "nginx")); err != nil {
		return err
	}

	storageDir, err := i.platformClient.InitStorage(App, App)
	if err != nil {
		return err
	}
	if err := i.GenerateConfig(storageDir); err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	return i.FixPermissions()
}

func (i *Installer) GenerateConfig(storageDir string) error {
	variables := Variables{
		App:          App,
		AppDir:       AppDir,
		DataDir:      DataDir,
		CommonDir:    CommonDir,
		StorageDir:   storageDir,
		Socket:       path.Join(DataDir, "audiobookshelf.sock"),
	}

	return config.Generate(
		path.Join(AppDir, "config"),
		path.Join(DataDir, "config"),
		variables,
	)
}

func (i *Installer) FixPermissions() error {
	if err := linux.Chown(DataDir, App); err != nil {
		return err
	}
	return linux.Chown(CommonDir, App)
}

func (i *Installer) BackupPreStop() error    { return i.PreRefresh() }
func (i *Installer) RestorePreStart() error  { return i.PostRefresh() }
func (i *Installer) RestorePostStart() error { return i.Configure() }
