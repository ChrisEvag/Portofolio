# 🔄 Chain Registry Auto-Update

Αυτόματη ενημέρωση του Cosmos Chain Registry **1 φορά την ημέρα**.

## 📋 Περιεχόμενα

- [Τι κάνει](#-τι-κάνει)
- [Χρήση](#-χρήση)
- [API Endpoints](#-api-endpoints)
- [Manual Execution](#️-manual-execution)
- [Configuration](#️-configuration)
- [Logs](#-logs)

---

## ✨ Τι κάνει

Το σύστημα:
1. ✅ **Κατεβάζει** το chain-registry από GitHub
2. ✅ **Ελέγχει** κάθε 24 ώρες για ενημερώσεις
3. ✅ **Ενημερώνει** αυτόματα τα asset lists
4. ✅ **Κρατάει μόνο** τις αλυσίδες που χρειάζεσαι (sparse checkout)
5. ✅ **Logάρει** όλες τις ενημερώσεις

---

## 🚀 Χρήση

### Αυτόματη Ενημέρωση

Το backend ξεκινάει αυτόματα τον updater:

```go
// main.go
chainRegistryUpdater := utils.NewChainRegistryUpdater()
chainRegistryUpdater.Start()
```

**Ροή:**
1. Τρέχει **αμέσως** όταν ξεκινάς το backend
2. Ελέγχει αν πέρασαν **24 ώρες**
3. Αν ναι → κάνει **git pull**
4. Επαναλαμβάνει κάθε **24 ώρες**

---

## 📡 API Endpoints

### 1. Κατάσταση Chain Registry
```bash
GET http://localhost:8080/api/chain-registry/status
```

**Response:**
```json
{
  "status": "ok",
  "last_update": "2025-10-18T10:30:00+03:00",
  "hours_since": 2.5,
  "needs_update": false
}
```

### 2. Αναγκαστική Ενημέρωση
```bash
POST http://localhost:8080/api/chain-registry/update
```

**Response:**
```json
{
  "status": "success",
  "message": "Chain registry ενημερώθηκε επιτυχώς",
  "last_update": "2025-10-18T13:00:00+03:00"
}
```

---

## 🛠️ Manual Execution

### PowerShell Script

```powershell
# Εκτέλεση του script απευθείας
.\backend\scripts\update-chain-registry.ps1
```

**Output:**
```
[2025-10-18 10:30:00] 🚀 Chain Registry Update Script Started
[2025-10-18 10:30:00] Repository: https://github.com/cosmos/chain-registry.git
[2025-10-18 10:30:00] Chains: osmosis
[2025-10-18 10:30:01] 📂 Το chain-registry υπάρχει - Ενημέρωση...
[2025-10-18 10:30:02] 🔍 Git pull...
[2025-10-18 10:30:03] ✅ Ενημέρωση ολοκληρώθηκε
[2025-10-18 10:30:03] 📊 Στατιστικά:
[2025-10-18 10:30:03]    osmosis - Assets: 120
[2025-10-18 10:30:03]    osmosis - Chain ID: osmosis-1
[2025-10-18 10:30:03] ✅ Ολοκλήρωση επιτυχής!
```

### Go Function

```go
import "portofoliov1/utils"

updater := utils.NewChainRegistryUpdater()

// Αναγκαστική ενημέρωση
err := updater.ForceUpdate()

// Έλεγχος τελευταίας ενημέρωσης
lastUpdate, err := updater.GetLastUpdateTime()
```

---

## ⚙️ Configuration

### Αλλαγή Αλυσίδων

**PowerShell Script:**
```powershell
# scripts/update-chain-registry.ps1
$CHAINS = @("osmosis", "cosmos", "juno")  # Πρόσθεσε εδώ
```

### Αλλαγή Συχνότητας

**Go Code:**
```go
// utils/chain_registry_updater.go
updateInterval: 12 * time.Hour, // Κάθε 12 ώρες αντί για 24
```

### Paths

```go
// Default paths
scriptPath:     "scripts/update-chain-registry.ps1"
lastUpdateFile: "data/chain-registry/.last_update"
```

---

## 📊 Logs

### Log File

```
backend/logs/chain-registry-update.log
```

**Περιεχόμενο:**
```
[2025-10-18 08:00:00] 🚀 Chain Registry Update Script Started
[2025-10-18 08:00:00] Τελευταία ενημέρωση: 2025-10-17 08:00:00 (24.0 ώρες πριν)
[2025-10-18 08:00:01] Πέρασαν πάνω από 24 ώρες - Χρειάζεται ενημέρωση
[2025-10-18 08:00:02] 📂 Το chain-registry υπάρχει - Ενημέρωση...
[2025-10-18 08:00:05] ✅ Ενημέρωση ολοκληρώθηκε
```

### Console Logs

Κατά την εκκίνηση του backend:
```
🔄 Έλεγχος chain-registry...
✅ Chain registry είναι ενημερωμένο
```

Κατά την ενημέρωση:
```
⏰ Χρόνος για ενημέρωση chain-registry...
📥 Ενημέρωση chain-registry...
📋 Script output:
   [...]
✅ Chain registry ενημερώθηκε επιτυχώς
```

---

## 🔍 Τι Κατεβάζει

```
data/chain-registry/
└── osmosis/
    ├── assetlist.json    # Assets & denoms
    ├── chain.json        # Chain info
    └── versions.json     # Version info
```

**assetlist.json** περιέχει:
- Token symbols (OSMO, ATOM, USDC...)
- IBC denoms (ibc/...)
- Logos & images
- Decimals & base denoms

---

## 🚨 Troubleshooting

### Git δεν βρέθηκε

```powershell
# Κατέβασε το Git
https://git-scm.com/download/win

# Εγκατάστασε και επανεκκίνησε το terminal
```

### Permission denied

```powershell
# Εκτέλεση με admin rights
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Script δεν τρέχει

```powershell
# Έλεγχος execution policy
Get-ExecutionPolicy

# Αλλαγή σε RemoteSigned
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Git pull αποτυγχάνει

```powershell
# Διαγραφή και επανεγκατάσταση
Remove-Item -Recurse -Force data/chain-registry
.\backend\scripts\update-chain-registry.ps1
```

---

## 🎯 Use Cases

### 1. Νέα Tokens
Όταν προστίθενται νέα tokens στο Osmosis, το chain-registry ενημερώνεται αυτόματα.

### 2. Logo Updates
Τα logos των tokens ενημερώνονται από το registry.

### 3. IBC Denoms
Νέα IBC paths προστίθενται αυτόματα.

### 4. Chain Upgrades
Μετά από upgrades, το chain.json ενημερώνεται με νέες πληροφορίες.

---

## 📝 Example: Manual Update via API

### cURL
```bash
# Έλεγχος status
curl http://localhost:8080/api/chain-registry/status

# Αναγκαστική ενημέρωση
curl -X POST http://localhost:8080/api/chain-registry/update
```

### JavaScript
```javascript
// Check status
fetch('http://localhost:8080/api/chain-registry/status')
  .then(res => res.json())
  .then(data => console.log(data));

// Force update
fetch('http://localhost:8080/api/chain-registry/update', {
  method: 'POST'
})
  .then(res => res.json())
  .then(data => console.log(data));
```

### PowerShell
```powershell
# Status
Invoke-RestMethod -Uri "http://localhost:8080/api/chain-registry/status"

# Update
Invoke-RestMethod -Uri "http://localhost:8080/api/chain-registry/update" -Method Post
```

---

## 📦 Files Created

```
backend/
├── scripts/
│   └── update-chain-registry.ps1    # PowerShell script
├── utils/
│   └── chain_registry_updater.go    # Go updater
├── logs/
│   └── chain-registry-update.log    # Update logs
└── data/
    └── chain-registry/
        ├── .last_update             # Timestamp file
        └── osmosis/
            ├── assetlist.json
            ├── chain.json
            └── versions.json
```

---

## ✅ Benefits

1. **Πάντα fresh data** - Νέα tokens εμφανίζονται αυτόματα
2. **Χωρίς manual work** - Τα πάντα γίνονται αυτόματα
3. **Ελαφρύ** - Sparse checkout, μόνο osmosis
4. **Reliable** - Auto-retry με git clone αν χρειαστεί
5. **Monitorable** - API endpoints για status & logs

---

**Made with ❤️ for automated DevOps**
