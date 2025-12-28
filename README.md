# STURDR-API
A REST-API for the SturDR software defined radio library.

## 1) Default Usage
There are two SQL tables created by this server, `navigation` and `satellites`. By default the `create`, `read`, `update`, and `delete` actions are set to their exact names after the table names.

- The `/satellite` and `/navigation` endpoints connect to specific rows of their respective table.
- To interact with the table you would simply send a request to the `/<table>/<action>` http address. For example:
    ```sh
    http://localhost:8000/satellite/create
    http://localhost:8000/navigation/read
    ```
- The `create`, `update`, and `delete` action will return a "success" string if the command was properly executed.
- The `read` action will return the extracted table data if properly executed.

### 1.1) Queries
The `create/read` and `update/delete` endpoints behave slightly differently. For simplicity, the `update/delete` actions have and appended `{id}` field to specify which sequency number you want to interact with:
```sh
http://localhost:8000/navigation/update/1
http://localhost:8000/navigation/delete/2
```
The `create/read` actions do not have this field.

There are a number of query parameters you can add to the end of each request including:
1) ***format*** (json or binary, default=json)
2) ***week*** (GPS week number)
3) ***tow*** (GPS time of week)
4) ***prn*** (Unique satellite PRN number)

These can be used to specify the in/out data format or to request specific data from the SQL tables. The "format" specifier can be used on any of the action endpoints. The "week" and "tow" specifiers are used to read data including and after the specified epoch. They can be used on any table. The "prn" specifier is used only on reads from the satellite table as a way to get data for only a single PRN. Example http strings may look like:
```sh
http://localhost:8000/navigation/create?format=json
http://localhost:8000/navigation/read?format=binary&week=2352&tow=507440.0
http://localhost:8000/satellite/read?format=json&week=2352&tow=507440.5&prn=0
```
Note, if the "week" and "tow" specifiers are excluded, the server will return only the latest epoch by default. 

### 1.2) Telemetry
The `/telemetry` endpoint will connect to all navigation and satellite data from the same sequence number, meaning it will return one navigation point and multiple satellite points at once. It is included to efficiently grab a bunch of data. All query rules additionally apply to the "telemetry" combined accessor. Example:
```sh
http://localhost:8000/telemetry/read?format=json&week=2352&tow=507440.0
```

### 1.3) GUI
The graphical user interface allows the user to visualize the data being processed by SturDR. It consists of two main webpages:
1. `http://{host}:{port}/`
The first link is the main "navigation view". It is the main hub containing most of the useful information and is located at the "/" address. It shows a leaflet map, prints navigation metrics, and plots satellite navigation metrics all in the same place.
2. `http://{host}:{port}/satellite-view`
The second link contains more detailed views into the satellite diagnostics. It is called the "satellite view" and is located at the "/satellite-view" address. To use it, a singular satellite must be chosen from the dropdown at the top. The page will then show four important plots about the chosen satellite: doppler, C/No, pseudorange, and correator histories.

There are buttons to easily switch between the two webpages.

## 2) Authors
1. Daniel Sturdivant (sturdivant20@gmail.com)

## 3) TODO
1. Force the SQL table to have a maximum size.
2. Allow user to select specific data to plot in the satellite view.
3. Make a github workflow.