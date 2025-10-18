# Quick Git Check Script
# Run this to verify what files will be pushed

Write-Host "`n🔍 Checking Git Status...`n" -ForegroundColor Cyan

# Check if git is available
try {
    $gitVersion = git --version
    Write-Host "✅ Git found: $gitVersion`n" -ForegroundColor Green
} catch {
    Write-Host "❌ Git not found in PATH" -ForegroundColor Red
    Write-Host "Please install Git or add it to PATH" -ForegroundColor Yellow
    exit 1
}

# Show current branch
Write-Host "📍 Current Branch:" -ForegroundColor Yellow
git branch --show-current
Write-Host ""

# Show status
Write-Host "📊 Git Status:" -ForegroundColor Yellow
git status --short
Write-Host ""

# Check if database files are ignored
Write-Host "🔒 Checking if database files are ignored..." -ForegroundColor Yellow
$dbFiles = @(
    "backend/data/database/osmosis_history.db",
    "backend/data/database/osmosis_history.db-wal",
    "backend/data/database/osmosis_history.db-shm",
    "backend/osmosis-tracker.exe"
)

foreach ($file in $dbFiles) {
    if (Test-Path $file) {
        $isIgnored = git check-ignore $file
        if ($isIgnored) {
            Write-Host "  ✅ $file - IGNORED (won't be pushed)" -ForegroundColor Green
        } else {
            Write-Host "  ⚠️  $file - NOT IGNORED (WILL be pushed!)" -ForegroundColor Red
        }
    }
}

Write-Host "`n📦 Files that WILL be committed:" -ForegroundColor Cyan
git status --short | Where-Object { $_ -match '^\s*[AM]' }

Write-Host "`n💡 To proceed:" -ForegroundColor Yellow
Write-Host "  1. git add ." -ForegroundColor White
Write-Host "  2. git commit -m 'Add SQLite storage with API endpoints'" -ForegroundColor White
Write-Host "  3. git push origin main (or your branch name)" -ForegroundColor White
Write-Host ""
