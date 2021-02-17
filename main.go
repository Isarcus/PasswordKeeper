package main

import (
	"fmt"
	"os"
	"strings"
	"zarks/system"
	"zarks/zmath/encrypt"
)

// FILENAME is what you think it is
const FILENAME = "PASSWORDS.enc"

// HEADER is what should be at the top of a decrypted file
const HEADER = "ONLY THE GREENEST BEAN"

type pair [2]string

func main() {
	var (
		r         = system.NewConsoleReader()
		data      = []byte{}
		password  string
		accList   map[string]pair
		otherPath string

		f *os.File
	)

	// Get path to file
	if system.FileExists(FILENAME) && system.QueryYN(r, "Would you like to open the existing PASSWORDS.enc file?") {
		f, _ = os.Open(FILENAME)
		defer f.Close()
	} else if system.QueryYN(r, "Would you like to create a new password file?") {
		for {
			password = system.Query(r, "Please enter your password.")
			if system.QueryYN(r, "Please confirm that "+password+" is your desired password.") {
				break
			}
		}
		goto decrypted // forgive me
	} else {
		for {
			otherPath = system.Query(r, "Please enter the path to the file you would like to open.")
			var err error
			f, err = os.Open(otherPath)
			if err == nil {
				defer f.Close()
				break
			}
		}
	}

	// Read the file
	for {
		buf := make([]byte, 256, 256)
		n, _ := f.Read(buf)

		data = append(data, buf...)

		if n < len(buf) {
			break
		}
	}
	f.Close()

	// Decrypt the file
	for {
		password = system.Query(r, "Please enter your decryption key.")
		if system.QueryYN(r, "Please confirm that "+password+" is your key") {
			aes := encrypt.NewAESCipher(password, encrypt.AES256)
			data = encrypt.Decrypt(aes, data)

			hdr := data[:len(HEADER)]
			if string(hdr) == HEADER {
				data = data[len(HEADER):]
			} else {
				fmt.Println("You appear to have entered the wrong password! Exiting now.")
				return
			}
			break
		}
	}

decrypted:

	// Parse the data
	accList = parse(data)
	fmt.Printf("You have %v saved accounts.\n", len(accList))

	// Now, the actual editor & shell!!!
	for {
		cmd := system.Query(r, "What would you like to do? (Enter H to list options)")
		cmd = strings.ToUpper(cmd)
		switch cmd[0] {
		case 'H':
			fmt.Println("* * * * * * * * * * * * * * * * *")
			fmt.Println("List of options:")
			fmt.Println("-> H: Display this help menu")
			fmt.Println("-> A: Add a new account")
			fmt.Println("-> D: Delete an existing account")
			fmt.Println("-> E: Edit an existing account")
			fmt.Println("-> L: List all account names")
			fmt.Println("-> P: Print a specific account's information")
			fmt.Println("-> Q: Save and quit this shell")
			fmt.Println("-> !: Change your encryption password")
			fmt.Println("* * * * * * * * * * * * * * * * *")

		case 'A':
			label := system.Query(r, "What would you like to call this account? (leave blank to cancel)")
			if label == "" {
				break
			}
			usn := system.Query(r, "What should the username be? (leave blank to cancel)")
			if usn == "" {
				break
			}
			pwd := system.Query(r, "What should the password be? (leave blank to cancel)")
			if pwd == "" {
				break
			}

			// test if current label is already used
			existingPair, ok := accList[label]
			if ok {
				fmt.Println("Oops, you're already using that label!")
				fmt.Println("-> Username:   ", existingPair[0])
				fmt.Println("-> Password:   ", existingPair[1])
				fmt.Println("If you're trying to edit an existing account, use the 'E' command.")
				break
			}

			// finally, append to accList if everything's alright
			accList[label] = pair{usn, pwd}

		case 'D':
			var (
				label   string
				thePair pair
				ok      bool
			)
			for {
				label = system.Query(r, "Please enter the name of the account you would like to delete. (leave blank to cancel)")
				if label == "" {
					ok = false
					break
				}

				thePair, ok = accList[label]
				if ok {
					break
				} else {
					fmt.Printf("Could not find account by the name of %v\n", label)
				}
			}

			if !ok {
				break
			}

			fmt.Println("You have selected the account by the name of:", label)
			fmt.Println("-> Username:   ", thePair[0])
			fmt.Println("-> Password:   ", thePair[1])
			if system.QueryYN(r, "Delete this account?") {
				delete(accList, label)
			}

		case 'E':
			var (
				label   string
				thePair pair
				ok      bool
			)
			for {
				label = system.Query(r, "Please enter the name of the account you would like to edit. (leave blank to cancel)")
				if label == "" {
					ok = false
					break
				}

				thePair, ok = accList[label]
				if ok {
					break
				} else {
					fmt.Printf("Could not find account by the name of %v\n", label)
				}
			}

			if !ok {
				break
			}

			fmt.Println("-> Username:   ", thePair[0])
			fmt.Println("-> Password:   ", thePair[1])
			fmt.Println()
			usn := system.Query(r, "What should the new username be? (leave blank for the same username)")
			if usn == "" {
				usn = thePair[0]
			}
			pwd := system.Query(r, "What should the new password be? (leave blank for the same password)")
			if pwd == "" {
				pwd = thePair[1]
			}
			accList[label] = pair{usn, pwd}
			fmt.Println("Successfully updated.")

		case 'L':
			fmt.Println("* * * * * * * * * *")
			for label := range accList {
				fmt.Println(label)
			}
			fmt.Println("* * * * * * * * * *")

		case 'P':
			label := system.Query(r, "Which account would you like to view?")
			if label == "" {
				break
			}
			thePair, ok := accList[label]
			if ok {
				fmt.Println("-> Username:   ", thePair[0])
				fmt.Println("-> Password:   ", thePair[1])
				fmt.Println("")
			} else {
				fmt.Println("Account", label, "not recognized.")
			}

		case 'Q':
			goto saveAndExit

		case '!':
			if !system.QueryYN(r, "Do you REALLY want to change your encryption key?") {
				break
			}

			var newKey string
			for {
				newKey = system.Query(r, "What would you like the new key to be?")
				if system.QueryYN(r, "Please confirm that you want the new key to be: \""+newKey+"\"") {
					break
				}
			}

			conf := system.Query(r, "One last step: Please type CONFIRM to confirm this change.")
			if conf == "CONFIRM" {
				password = newKey
				fmt.Println("Password successfully changed!")
			} else {
				fmt.Println("Password change aborted!")
			}
		}
	}
saveAndExit:

	var saveData = []byte(HEADER)
	for label, thePair := range accList {
		saveData = append(saveData, []byte(label)...)
		saveData = append(saveData, '\n')
		saveData = append(saveData, []byte(thePair[0])...)
		saveData = append(saveData, '\n')
		saveData = append(saveData, []byte(thePair[1])...)
		saveData = append(saveData, '\n')
		saveData = append(saveData, '\n')
	}
	aes := encrypt.NewAESCipher(password, encrypt.AES256)
	saveData = encrypt.Encrypt(aes, saveData)

	// Now, remove and rewrite the password file

	var path string
	if otherPath == "" {
		path = FILENAME
	} else {
		path = otherPath
	}
	os.Remove(path)
	fs := system.CreateFile(path)
	fs.Write(saveData)

	fmt.Println("File saved at", path)
}

func parse(data []byte) map[string]pair {
	const (
		findLbl = iota
		findUsn
		findPwd
	)
	var (
		pwds    = make(map[string]pair)
		currLbl = ""
		currUsn = ""
		currPwd = ""

		finding = findLbl
	)

	for _, b := range data {
		// if you've reached the null-repeat end of the file
		if b == '\000' {
			break
		}

		switch finding { // this is truly awful, why did I do it this way
		case findLbl:
			if b == '\n' {
				if len(currLbl) > 0 {
					finding = findUsn
				}
			} else {
				currLbl += string(b)
			}

		case findUsn:
			if b == '\n' {
				if len(currUsn) > 0 {
					finding = findPwd
				}
			} else {
				currUsn += string(b)
			}

		case findPwd:
			if b == '\n' {
				if len(currPwd) > 0 {
					pwds[currLbl] = pair{currUsn, currPwd}
				} else {
					fmt.Println("Incorrect formatting!", currLbl, currUsn)
				}
				finding = findLbl
				currLbl = ""
				currUsn = ""
				currPwd = ""
			} else {
				currPwd += string(b)
			}
		}
	}
	return pwds
}
