@echo off
setlocal enabledelayedexpansion

:: ACC Server Manager Deployment Script
:: This is a wrapper for the PowerShell deployment script

echo.
echo ===================================
echo ACC Server Manager Deployment Tool
echo ===================================
echo.

:: Check if PowerShell is available
where powershell >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: PowerShell is not installed or not in PATH
    echo Please install PowerShell to use this deployment script
    pause
    exit /b 1
)

:: Check if required parameters are provided
if "%~1"=="" (
    echo Usage: deploy.bat [ServerHost] [ServerUser] [DeployPath] [additional options]
    echo.
    echo Examples:
    echo   deploy.bat 192.168.1.100 admin "C:\AccServerManager"
    echo   deploy.bat server.example.com deploy "C:\Services\AccServerManager" -SkipTests
    echo   deploy.bat 192.168.1.100 admin "C:\AccServerManager" -WhatIf
    echo.
    echo For more options, run: powershell -File scripts\deploy.ps1
    pause
    exit /b 1
)

:: Set variables
set SERVER_HOST=%~1
set SERVER_USER=%~2
set DEPLOY_PATH=%~3

:: Shift parameters to get additional options
shift
shift
shift

set ADDITIONAL_PARAMS=
:loop
if "%~1"=="" goto :continue
set ADDITIONAL_PARAMS=%ADDITIONAL_PARAMS% %1
shift
goto :loop
:continue

echo Server Host: %SERVER_HOST%
echo Server User: %SERVER_USER%
echo Deploy Path: %DEPLOY_PATH%
if not "%ADDITIONAL_PARAMS%"=="" (
    echo Additional Parameters: %ADDITIONAL_PARAMS%
)
echo.

:: Confirm deployment
set /p CONFIRM=Do you want to proceed with deployment? (Y/N):
if /i not "%CONFIRM%"=="Y" (
    echo Deployment cancelled.
    pause
    exit /b 0
)

echo.
echo Starting deployment...
echo.

:: Execute PowerShell deployment script
powershell -ExecutionPolicy Bypass -File "%~dp0deploy.ps1" -ServerHost "%SERVER_HOST%" -ServerUser "%SERVER_USER%" -DeployPath "%DEPLOY_PATH%" %ADDITIONAL_PARAMS%

if %errorlevel% neq 0 (
    echo.
    echo Deployment failed with error code: %errorlevel%
    pause
    exit /b %errorlevel%
)

echo.
echo Deployment completed successfully!
pause
