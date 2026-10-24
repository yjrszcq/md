cd .\web
call npm install
call npm run build
cd ..\
rmdir /s/q .\md\web\
xcopy /s/e .\web\dist\ .\md\web\
cd .\md
if exist "%USERPROFILE%\sdk\go1.27.3\bin\go.exe" (
  call "%USERPROFILE%\sdk\go1.27.3\bin\go.exe" build
) else (
  echo Go 1.27.3 SDK not found: %USERPROFILE%\sdk\go1.27.3\bin\go.exe
  exit /b 1
)
echo md build finished
pause
