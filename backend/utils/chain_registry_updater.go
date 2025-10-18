package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ChainRegistryUpdater - Διαχειριστής ενημέρωσης chain registry
type ChainRegistryUpdater struct {
	scriptPath     string
	updateInterval time.Duration
	lastUpdateFile string
	isRunning      bool
	stopChan       chan bool
}

// NewChainRegistryUpdater - Δημιουργία νέου updater
func NewChainRegistryUpdater() *ChainRegistryUpdater {
	scriptPath := filepath.Join("scripts", "update-chain-registry-full.ps1")
	lastUpdateFile := filepath.Join("data", "chain-registry", ".last_update")

	return &ChainRegistryUpdater{
		scriptPath:     scriptPath,
		updateInterval: 7 * 24 * time.Hour,
		lastUpdateFile: lastUpdateFile,
		stopChan:       make(chan bool),
	}
}

// Start - Εκκίνηση αυτόματης ενημέρωσης
func (u *ChainRegistryUpdater) Start() error {
	if u.isRunning {
		return fmt.Errorf("updater ήδη τρέχει")
	}

	u.isRunning = true

	// Τρέξε αμέσως την πρώτη φορά
	log.Println("🔄 Έλεγχος chain-registry...")
	if err := u.RunUpdate(); err != nil {
		log.Printf("⚠️  Προειδοποίηση: Αποτυχία ενημέρωσης chain-registry: %v", err)
	}

	// Ξεκίνα background goroutine για ενημερώσεις
	go u.updateLoop()

	return nil
}

// Stop - Διακοπή αυτόματης ενημέρωσης
func (u *ChainRegistryUpdater) Stop() {
	if !u.isRunning {
		return
	}

	u.isRunning = false
	u.stopChan <- true
}

// updateLoop - Loop για αυτόματες ενημερώσεις
func (u *ChainRegistryUpdater) updateLoop() {
	ticker := time.NewTicker(u.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("⏰ Χρόνος για ενημέρωση chain-registry...")
			if err := u.RunUpdate(); err != nil {
				log.Printf("⚠️  Αποτυχία ενημέρωσης: %v", err)
			}
		case <-u.stopChan:
			log.Println("🛑 Chain registry updater σταμάτησε")
			return
		}
	}
}

// RunUpdate - Εκτέλεση του PowerShell script
func (u *ChainRegistryUpdater) RunUpdate() error {
	// Έλεγχος αν χρειάζεται ενημέρωση
	if !u.needsUpdate() {
		log.Println("✅ Chain registry είναι ενημερωμένο")
		return nil
	}

	log.Println("📥 Ενημέρωση chain-registry...")

	// Έλεγχος αν υπάρχει το script
	if _, err := os.Stat(u.scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("το script δεν βρέθηκε: %s", u.scriptPath)
	}

	// Εκτέλεση PowerShell script
	cmd := exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", u.scriptPath)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("σφάλμα εκτέλεσης script: %v\nOutput: %s", err, string(output))
	}

	log.Printf("📋 Script output:\n%s", string(output))
	log.Println("✅ Chain registry ενημερώθηκε επιτυχώς")

	return nil
}

// needsUpdate - Έλεγχος αν χρειάζεται ενημέρωση
func (u *ChainRegistryUpdater) needsUpdate() bool {
	// Αν δεν υπάρχει το αρχείο, χρειάζεται ενημέρωση
	if _, err := os.Stat(u.lastUpdateFile); os.IsNotExist(err) {
		return true
	}

	// Διάβασε την τελευταία ενημέρωση
	data, err := os.ReadFile(u.lastUpdateFile)
	if err != nil {
		log.Printf("⚠️  Δεν μπόρεσε να διαβάσει last_update: %v", err)
		return true
	}

	// Clean the string (remove \r\n)
	timestampStr := string(data)
	timestampStr = strings.TrimSpace(timestampStr)

	lastUpdate, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		log.Printf("⚠️  Δεν μπόρεσε να parse την ημερομηνία: %v", err)
		return true
	}

	// Αν πέρασαν πάνω από 7 ημέρες (1 εβδομάδα)
	hoursSince := time.Since(lastUpdate).Hours()
	return hoursSince >= (7 * 24)
}

// ForceUpdate - Αναγκαστική ενημέρωση (αγνοώντας το 24ωρο)
func (u *ChainRegistryUpdater) ForceUpdate() error {
	log.Println("🔄 Αναγκαστική ενημέρωση chain-registry...")

	// Διαγραφή του last_update file για να αναγκάσουμε update
	if err := os.Remove(u.lastUpdateFile); err != nil && !os.IsNotExist(err) {
		log.Printf("⚠️  Δεν μπόρεσε να διαγράψει last_update: %v", err)
	}

	return u.RunUpdate()
}

// GetLastUpdateTime - Επιστρέφει την τελευταία ενημέρωση
func (u *ChainRegistryUpdater) GetLastUpdateTime() (time.Time, error) {
	if _, err := os.Stat(u.lastUpdateFile); os.IsNotExist(err) {
		return time.Time{}, fmt.Errorf("δεν έχει γίνει ενημέρωση ακόμα")
	}

	data, err := os.ReadFile(u.lastUpdateFile)
	if err != nil {
		return time.Time{}, err
	}

	// Clean the string (remove \r\n)
	timestampStr := string(data)
	timestampStr = strings.TrimSpace(timestampStr)

	return time.Parse("2006-01-02 15:04:05", timestampStr)
}
