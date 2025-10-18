# ====================================================================
# Chain Registry FULL Download Script (No Git Required)
# ====================================================================
# Downloads ALL chains from GitHub for future expansion
# ====================================================================

$ErrorActionPreference = "Stop"

# Paths
$BACKEND_DIR = Split-Path -Parent $PSScriptRoot
$DATA_DIR = Join-Path $BACKEND_DIR "data"
$CHAIN_REGISTRY_DIR = Join-Path $DATA_DIR "chain-registry"
$LOGS_DIR = Join-Path $BACKEND_DIR "logs"
$LOG_FILE = Join-Path $LOGS_DIR "chain-registry-update.log"
$LAST_UPDATE_FILE = Join-Path $CHAIN_REGISTRY_DIR ".last_update"

# GitHub Configuration
$GITHUB_API_URL = "https://api.github.com/repos/cosmos/chain-registry/contents"
$GITHUB_RAW_URL = "https://raw.githubusercontent.com/cosmos/chain-registry/master"

# Files to download for each chain
$FILES_TO_DOWNLOAD = @(
    "assetlist.json",
    "chain.json"
)

# Optional files (won't fail if missing)
$OPTIONAL_FILES = @(
    "versions.json",
    "ibc_assets.json"
)

# ====================================================================
# Functions
# ====================================================================

function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "[$timestamp] $Message"
    Write-Host $logMessage
    
    if (-not (Test-Path $LOGS_DIR)) {
        New-Item -ItemType Directory -Path $LOGS_DIR -Force | Out-Null
    }
    
    Add-Content -Path $LOG_FILE -Value $logMessage
}

function Should-Update {
    if (-not (Test-Path $LAST_UPDATE_FILE)) {
        Write-Log "No last update file found"
        return $true
    }
    
    try {
        $lastUpdate = Get-Content $LAST_UPDATE_FILE -Raw -ErrorAction Stop
        $lastUpdate = $lastUpdate.Trim()
        $lastUpdateDate = [DateTime]::Parse($lastUpdate)
        $now = Get-Date
        $hoursSinceUpdate = ($now - $lastUpdateDate).TotalHours
        
        Write-Log "Last update: $lastUpdate ($([math]::Round($hoursSinceUpdate, 2)) hours ago)"
        
        if ($hoursSinceUpdate -ge 24) {
            Write-Log "More than 24 hours passed - Update needed"
            return $true
        }
        
        Write-Log "No update needed yet"
        return $false
    }
    catch {
        Write-Log "Error reading last update file: $_"
        return $true
    }
}

function Get-AllChains {
    Write-Log "Fetching list of all chains from GitHub..."
    
    try {
        $response = Invoke-RestMethod -Uri $GITHUB_API_URL -Headers @{
            "User-Agent" = "PowerShell-ChainRegistry-Updater"
        }
        
        # Filter only directories (chains)
        $chains = $response | Where-Object { $_.type -eq "dir" } | Select-Object -ExpandProperty name
        
        # Exclude special folders
        $excludeFolders = @(".github", "_IBC", "_non-cosmos", "testnets", ".vscode")
        $chains = $chains | Where-Object { $excludeFolders -notcontains $_ }
        
        Write-Log "Found $($chains.Count) chains to process"
        
        return $chains
    }
    catch {
        Write-Log "ERROR - Failed to fetch chain list: $_"
        throw
    }
}

function Download-ChainFiles {
    param(
        [string]$Chain
    )
    
    # Create chain directory
    $chainDir = Join-Path $CHAIN_REGISTRY_DIR $Chain
    if (-not (Test-Path $chainDir)) {
        New-Item -ItemType Directory -Path $chainDir -Force | Out-Null
    }
    
    $successCount = 0
    $failCount = 0
    
    # Required files
    foreach ($file in $FILES_TO_DOWNLOAD) {
        try {
            $url = "$GITHUB_RAW_URL/$Chain/$file"
            $destination = Join-Path $chainDir $file
            
            Invoke-WebRequest -Uri $url -OutFile $destination -ErrorAction Stop | Out-Null
            
            # Verify it's valid JSON
            $content = Get-Content $destination -Raw | ConvertFrom-Json | Out-Null
            
            $successCount++
        }
        catch {
            $failCount++
            # Required files must succeed
            throw "Failed to download required file $file for $Chain"
        }
    }
    
    # Optional files
    foreach ($file in $OPTIONAL_FILES) {
        try {
            $url = "$GITHUB_RAW_URL/$Chain/$file"
            $destination = Join-Path $chainDir $file
            
            Invoke-WebRequest -Uri $url -OutFile $destination -ErrorAction Stop | Out-Null
            $content = Get-Content $destination -Raw | ConvertFrom-Json | Out-Null
            
            $successCount++
        }
        catch {
            # Optional files can fail silently
        }
    }
    
    return $successCount
}

function Update-AllChains {
    Write-Log "======================================"
    Write-Log "Chain Registry FULL Update (No Git)"
    Write-Log "======================================"
    
    # Create directories
    if (-not (Test-Path $DATA_DIR)) {
        New-Item -ItemType Directory -Path $DATA_DIR -Force | Out-Null
    }
    
    if (-not (Test-Path $CHAIN_REGISTRY_DIR)) {
        New-Item -ItemType Directory -Path $CHAIN_REGISTRY_DIR -Force | Out-Null
    }
    
    # Get all chains
    $chains = Get-AllChains
    
    $totalFiles = 0
    $successfulChains = 0
    $failedChains = 0
    $processed = 0
    
    foreach ($chain in $chains) {
        $processed++
        
        try {
            Write-Log "[$processed/$($chains.Count)] Processing $chain..."
            $filesDownloaded = Download-ChainFiles -Chain $chain
            $totalFiles += $filesDownloaded
            $successfulChains++
            Write-Log "   SUCCESS - $chain ($filesDownloaded files)"
        }
        catch {
            $failedChains++
            Write-Log "   FAILED - $chain - $_"
        }
        
        # Progress indicator every 10 chains
        if ($processed % 10 -eq 0) {
            Write-Log "Progress: $processed/$($chains.Count) chains processed..."
        }
    }
    
    Write-Log ""
    Write-Log "======================================"
    Write-Log "Update Summary:"
    Write-Log "   Total Chains: $($chains.Count)"
    Write-Log "   Successful: $successfulChains"
    Write-Log "   Failed: $failedChains"
    Write-Log "   Total Files: $totalFiles"
    Write-Log "======================================"
    
    if ($successfulChains -eq 0) {
        throw "No chains were downloaded successfully"
    }
}

function Update-Timestamp {
    $now = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    
    $lastUpdateDir = Split-Path $LAST_UPDATE_FILE
    if (-not (Test-Path $lastUpdateDir)) {
        New-Item -ItemType Directory -Path $lastUpdateDir -Force | Out-Null
    }
    
    Set-Content -Path $LAST_UPDATE_FILE -Value $now -NoNewline
    Write-Log "Updated timestamp: $now"
}

function Show-Statistics {
    Write-Log ""
    Write-Log "Statistics for key chains:"
    Write-Log "======================================"
    
    $keyChains = @("osmosis", "cosmos", "juno", "stargaze", "akash", "celestia")
    
    foreach ($chain in $keyChains) {
        $chainPath = Join-Path $CHAIN_REGISTRY_DIR $chain
        
        if (Test-Path $chainPath) {
            $assetlistPath = Join-Path $chainPath "assetlist.json"
            $chainJsonPath = Join-Path $chainPath "chain.json"
            
            if (Test-Path $assetlistPath) {
                try {
                    $assetlist = Get-Content $assetlistPath -Raw | ConvertFrom-Json
                    $assetCount = $assetlist.assets.Count
                    Write-Log "   $chain - Assets: $assetCount"
                }
                catch {
                    Write-Log "   $chain - Could not read assetlist"
                }
            }
            
            if (Test-Path $chainJsonPath) {
                try {
                    $chainInfo = Get-Content $chainJsonPath -Raw | ConvertFrom-Json
                    Write-Log "   $chain - Chain ID: $($chainInfo.chain_id)"
                }
                catch {
                    Write-Log "   $chain - Could not read chain info"
                }
            }
        }
    }
    
    Write-Log "======================================"
}

function Test-InternetConnection {
    try {
        $response = Invoke-WebRequest -Uri "https://api.github.com" -Method Head -TimeoutSec 5 -ErrorAction Stop
        return $true
    }
    catch {
        Write-Log "ERROR - No internet connection or GitHub is unreachable"
        return $false
    }
}

# ====================================================================
# Main Script
# ====================================================================

try {
    Write-Log ""
    Write-Log "Chain Registry FULL Download Started (No Git Required)"
    Write-Log "Source: GitHub API + Raw URLs"
    Write-Log "Downloading: ALL Cosmos chains"
    
    # Check internet connection
    if (-not (Test-InternetConnection)) {
        exit 1
    }
    
    # Check if we need to update
    if (-not (Should-Update)) {
        Write-Log "Skipping update"
        exit 0
    }
    
    # Download all chains
    Update-AllChains
    
    # Update timestamp
    Update-Timestamp
    
    # Show statistics
    Show-Statistics
    
    Write-Log "SUCCESS - Full chain registry update completed!"
    Write-Log ""
    
    exit 0
}
catch {
    Write-Log "ERROR: $_"
    Write-Log $_.ScriptStackTrace
    exit 1
}
