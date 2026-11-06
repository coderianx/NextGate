package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/coderianx/NextGate/internal/admin"
	"github.com/coderianx/NextGate/internal/admin/password"
	"github.com/coderianx/NextGate/internal/config"
	"github.com/coderianx/NextGate/internal/store"
)

func main() {

	adminServerPort := flag.Int("admin-port", 9090, "Admin server port for start admin server")
	setAdminPassword := flag.Bool("admin-password", false, "Set admin password (sudo required)")

	flag.Parse()

	if *setAdminPassword {
		if os.Geteuid() != 0 {
			fmt.Println("sudo is required please use with sudo")
			os.Exit(1)
		}

		var adminPassword string

		fmt.Print("Enter New Admin Password: ")
		fmt.Scan(&adminPassword)

		err := password.SetAdminPassword(adminPassword)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Admin password updated.")
	}

	haveAdminPasswd, err := password.CheckAdminPassword()

	if err != nil {
		log.Fatal(err)
	}

	if !haveAdminPasswd {
		fmt.Println("Please set admin password!")
	}

	err = config.CheckEnvs()

	if err != nil {
		log.Fatal(err)
	}

	err = config.CreateAESKey()

	if err != nil {
		log.Fatal(err)
	}

	err = store.ConnectDB(
		filepath.Join(
			config.GetNextGatePath(),
			"nextgate.db",
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	defer store.CloseDB()

	err = store.CreateTables()

	if err != nil {
		log.Fatal(err)
	}

	adminServer := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", *adminServerPort),
		Handler: admin.NewRouter(),

		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Admin server running on 127.0.0.1:%d", *adminServerPort)

	if err := adminServer.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
