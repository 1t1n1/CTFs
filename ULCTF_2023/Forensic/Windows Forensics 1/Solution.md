1. Open Security logs in Event Viewer (Windows)
2. Filter on EventID 4625 (Failed log on)
3. Remember the timestamp and remove the filter. Go to the location of the log by looking at the timestamp
4. By looking at the following logs, you can find the correct one which corresponds to the successful login. It's the first with EventID 4624.
5. Flag is `ULCTF-2023-03-24T02:02:10.8306207Z`