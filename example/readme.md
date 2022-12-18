# Example Scripts for Testing the Library

The `example` folder contains a collection of shell scripts that demonstrate how to use the library to perform various tasks that required OAuth user login.

## Running the read_a_sheet.sh Script

To run the `read_a_sheet.sh` script, follow these steps:

1. Make sure you have the following installed on your system:
   - curl
   - jq
   - column
   - gisha (if you want to use the `gisha login` command to obtain an access token)
2. Open a terminal and navigate to the `example` folder.
3. Run the following command:
```
./read_a_sheet.sh {spreadsheetId} {sheetName}
```

Replace `{spreadsheetId}` and `{sheetName}` with the appropriate values for your spreadsheet. The script will retrieve all the data from the first sheet in the spreadsheet and print it as a table.

If you want to specify a different range of cells to retrieve, you can pass it as an argument:

```
./read_a_sheet.sh {spreadsheetId} {sheetName} {range}
```

Replace `{spreadsheetId}`, `{sheetName}`, and `{range}` with the appropriate values for your spreadsheet and range.