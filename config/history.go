package config

//GetPathHistory returns the recent DB paths that have been opened, if none returns [""]
func (conf Config) GetPathHistory() []string {
	return conf.History
}

//AddToPathHistory Add a db file path to the history
func (conf *Config) AddToPathHistory(path string) error {
	newHistory := make([]string, 1)
	newHistory[0] = path
	// Add any unique already in the history
	for _, entry := range conf.History {
		if entry != path {
			newHistory = append(newHistory, entry)
		}
	}

	if len(newHistory) > conf.HistoryLength {
		conf.History = newHistory[:conf.HistoryLength]
	} else {
		conf.History = newHistory
	}
	err := conf.Save()
	if err != nil {
		return err
	}
	return nil
}
