#!/bin/bash

# Set the variables from the script arguments
spreadsheet_id=$1
sheet_name=$2
range=${3:-A1:Z999}

# Run the gisha login command to obtain the access token
access_token=$(./cmd login -c "" -s "" -u https://accounts.google.com/o/oauth2/v2/auth -t https://oauth2.googleapis.com/token -S https://www.googleapis.com/auth/spreadsheets.readonly)

# Construct the URL for the API request
url="https://sheets.googleapis.com/v4/spreadsheets/$spreadsheet_id/values/$sheet_name!$range"

# Make the request to the API
response=$(curl -H "Authorization: Bearer $access_token" "$url")

# Extract the values from the JSON response and print them as a table
echo "$response" | jq -r '.values[] | @tsv' | column -t -s $'\t'