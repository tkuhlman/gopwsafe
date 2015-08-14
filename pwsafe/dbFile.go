package pwsafe

import "os"

//OpenPWSafeFile Opens a password safe v3 file and decrypts with the supplied password
func OpenPWSafeFile(dbPath string, passwd string) (DB, error) {
	var db V3

	// Open the file
	f, err := os.Open(dbPath)
	if err != nil {
		return &db, err
	}
	defer f.Close()

	_, err = db.Decrypt(f, passwd)

	return &db, err
}
