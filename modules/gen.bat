@echo off
setlocal

:: Get the module name
set /p module_name="Please enter the module name: "

:: Check if the module name is empty
if "%module_name%"=="" (
    echo Error: Module name cannot be empty.
    exit /b 1
)

:: Create the module directory
mkdir %module_name%
if errorlevel 1 (
    echo Error: Could not create directory %module_name%.
    exit /b 1
)

:: Change to the module directory
cd %module_name%

:: Create main.tf file
(
    echo # Terraform Main Configuration
    echo provider "aws" {
    echo   region = "us-east-1"
    echo }
) > main.tf

:: Create variables.tf file
(
    echo # Variable Definitions
    echo variable "example_variable" {
    echo   description = "Example variable"
    echo   type        = string
    echo }
) > variables.tf

:: Create outputs.tf file
(
    echo # Output Definitions
    echo output "example_output" {
    echo   value = "Example output"
    echo }
) > outputs.tf

:: Create README.md file
(
    echo # %module_name% Module
    echo This is the documentation for the %module_name% module.
) > README.md

echo Files created successfully!

endlocal
pause