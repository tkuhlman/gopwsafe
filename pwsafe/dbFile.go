package pwsafe

import "os"

func OpenPWSafeFile(dbPath string, passwd string) (DB, error) {
	var db PWSafeV3

	// Open the file
	f, err := os.Open(dbPath)
	if err != nil {
		return &db, err
	}
	defer f.Close()

	_, err = db.Decrypt(f, passwd)

	return &db, err
}
