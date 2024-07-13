This was the second Defence challenge of CyberSci regionals 2023. The goal was to find the 2 vulnerabilities and fix them. During the competition, it was possible to as a bot to try the attack and see if it still works. Unfortunately, it's not possible now since we don't have the bot

Solution 1: There was a SQLi possible in DescribeItem. 
Bad Line: `err := db.QueryRow("SELECT * FROM ITEMS WHERE id = '"+itemId+"'").Scan(&item.Id, &item.Name, &item.Description, &item.ImageUrl, &item.Price)`
Corrected: `err := db.QueryRow("SELECT * FROM ITEMS WHERE id = ?", itemId).Scan(&item.Id, &item.Name, &item.Description, &item.ImageUrl, &item.Price)`

Solution 2: There was a problem with the validation of the authentication because the coders reimplemented the verification instead of using the library's built in functions. Unfortunately, I wasn't able to complete the challenge in time so I don't have the exact solution
