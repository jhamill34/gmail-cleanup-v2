## Usage 

### Step 0

Get your credentials from the [google cloud console](https://console.cloud.google.com/)
and store them in `credentials.json`

### Step 1

Get a list of emails in your inbox. This will require auth if this 
is the first time running the script. 

```bash 
go run ./cmd/fetch/main.go
```

### Step 2

Convert your sender list to a "delete list" 

```bash 
cat senders.json | jq -r 'keys[] as $k | "\($k) \(.[$k] | length)"' > delete_list.txt
```

### Step 3

Go through your `delete_list.txt` file and add `#` at the beginning of the 
line for any email senders you don't want to delete.


### Step 4

Run the delete script. If you just ran `fetch.go` you might need to delete the `token.json` 
file since this script requires different scopes. 

```bash 
go run ./cmd/delete/main.go
```

This will add a `!` to the beginning of any line that was successfully deleted 
(this acts as a `#` and will be ignored in future runs if you want to iteratively delete emails.)


