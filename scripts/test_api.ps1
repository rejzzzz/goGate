$baseUrl = if ($env:TARGET_URL) { $env:TARGET_URL } else { "https://api.gogate.rejwanul.dev" }
$apiKey = if ($env:TEST_API_KEY) { $env:TEST_API_KEY } else { 
    Write-Host "Error: TEST_API_KEY environment variable is not set." -ForegroundColor Red
    exit 1
}
$headers = @{ "X-API-Key" = $apiKey }
$ErrorActionPreference = "Stop"

try {
    Write-Host "=== Testing Gateway Health ===" -ForegroundColor Cyan
    $health = Invoke-RestMethod -Uri "$baseUrl/health" -Method Get -Headers $headers
    $health | ConvertTo-Json
    
    Write-Host "`n=== Testing Users Endpoint ===" -ForegroundColor Cyan
    $users = Invoke-RestMethod -Uri "$baseUrl/api/v1/users" -Method Get -Headers $headers
    $users | ConvertTo-Json
    
    Write-Host "`n=== Testing Orders Endpoint ===" -ForegroundColor Cyan
    $orders = Invoke-RestMethod -Uri "$baseUrl/api/v1/orders" -Method Get -Headers $headers
    $orders | ConvertTo-Json
    
    Write-Host "`nAll tests completed successfully!" -ForegroundColor Green
} catch {
    Write-Host "`nError occurred during testing:" -ForegroundColor Red
    Write-Host $_.Exception.Message
}
