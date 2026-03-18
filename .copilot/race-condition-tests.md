# Command Serialization Race Condition Test

## Test Scripts for Command Concurrency

### Test 1: Sequential Sync Commands (Should be serialized)
```powershell
# Run these commands quickly to test serialization
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return "Command 1: " .. os.epoch("utc")'
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return "Command 2: " .. os.epoch("utc")'
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return "Command 3: " .. os.epoch("utc")'
```

### Test 2: Parallel Sync Commands (PowerShell Jobs)
```powershell
# Start multiple commands simultaneously
$job1 = Start-Job { Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return "Job 1: " .. os.epoch("utc")' }
$job2 = Start-Job { Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return "Job 2: " .. os.epoch("utc")' }
$job3 = Start-Job { Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return "Job 3: " .. os.epoch("utc")' }

# Wait for all jobs and get results
$results = @($job1, $job2, $job3) | Wait-Job | Receive-Job
$results | ForEach-Object { Write-Host $_.result.result }

# Clean up jobs
Remove-Job $job1, $job2, $job3
```

### Test 3: Mixed Sync/Async Commands
```powershell
# Test mixing sync and async modes
$asyncJob = Start-Job { Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Headers @{"X-Execution-Mode"="async"} -Body 'print("Async: " .. os.epoch("utc"))' }
$syncResult = Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return "Sync: " .. os.epoch("utc")'

Write-Host "Sync result: $($syncResult.result.result)"
$asyncResult = $asyncJob | Wait-Job | Receive-Job
Write-Host "Async result: $($asyncResult.message)"
Remove-Job $asyncJob
```

### Test 4: Movement Commands Serialization
```powershell
# Test that movement commands are properly serialized
$moveJob1 = Start-Job { Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'local success = turtle.forward(); return "Move 1: " .. tostring(success)' }
$moveJob2 = Start-Job { Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'local success = turtle.turnRight(); return "Turn 1: " .. tostring(success)' }
$moveJob3 = Start-Job { Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'local success = turtle.back(); return "Move 2: " .. tostring(success)' }

$moveResults = @($moveJob1, $moveJob2, $moveJob3) | Wait-Job | Receive-Job
$moveResults | ForEach-Object { Write-Host $_.result.result }
Remove-Job $moveJob1, $moveJob2, $moveJob3
```

## Expected Results with Serialization Fix

### Before Fix (Race Conditions)
- Commands could execute simultaneously
- Unpredictable execution order
- Potential conflicts between movement commands

### After Fix (Serialized)
- All commands execute in FIFO order (first received, first executed)
- Sync and async commands are both serialized per turtle
- No race conditions between concurrent API calls
- Server logs show "Acquired lock" and "Released lock" messages

## Verification

### Server Log Patterns to Look For:
```
[Command Serialization] Created command serialization for turtle 0
[Command Serialization] Acquired lock for turtle 0
[Sync Command (Default)]: [ 0 ] return "Command 1: " .. os.epoch("utc")
[Command Serialization] Released lock for turtle 0
[Command Serialization] Acquired lock for turtle 0
[Sync Command (Default)]: [ 0 ] return "Command 2: " .. os.epoch("utc")
[Command Serialization] Released lock for turtle 0
```

### Expected Behavior:
1. **FIFO Order**: Commands execute in the exact order they were received
2. **No Overlapping**: Each command completes before the next starts
3. **Mixed Mode Support**: Sync and async commands are serialized together
4. **Per-Turtle Isolation**: Commands for turtle 0 don't block commands for turtle 1

## Manual Testing Steps

1. **Start Server**: Ensure the updated server is running
2. **Connect Turtle**: Restart turtle program in ComputerCraft to reconnect
3. **Run Tests**: Execute the test scripts above
4. **Check Logs**: Verify serialization messages in server output
5. **Verify Order**: Confirm commands execute in sent order

## Success Criteria

✅ **Commands execute in FIFO order**
✅ **No race conditions between simultaneous requests**  
✅ **Server logs show proper lock acquisition/release**
✅ **Both sync and async modes work correctly**
✅ **Performance impact is minimal (sub-millisecond locking)**
