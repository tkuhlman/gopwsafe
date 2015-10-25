package config

//todo implement and likely these will be methods of a config interface

//GetPathHistory returns the recent DB paths that have been opened, if none returns [""]
func (conf Config) GetPathHistory() []string {
	return conf.History
}

//AddToPathHistory Add a db file path to the history
func (conf *Config) AddToPathHistory(path string) error {
	//todo handle duplicates and handle only keeping a certain amount of history
	conf.History = append(conf.History, path)
	err := conf.Save()
	if err != nil {
		return err
	}
	return nil
}
