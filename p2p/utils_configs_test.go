package p2p

func createTestingGenesisAndConfigFiles(t *testing.T, cfg modules.Config, genesisState modules.GenesisState, n int) {
	config, err := json.Marshal(cfg.P2P)
	require.NoError(t, err)

	genesis, err := json.Marshal(genesisState.ConsensusGenesisState)
	require.NoError(t, err)

	genesisFile := make(map[string]json.RawMessage)
	configFile := make(map[string]json.RawMessage)
	moduleName := new(p2pModule).GetModuleName()

	genesisFile[test_artifacts.GetGenesisFileName(moduleName)] = genesis
	configFile[moduleName] = config
	genesisFileBz, err := json.MarshalIndent(genesisFile, "", "    ")
	require.NoError(t, err)

	p2pFileBz, err := json.MarshalIndent(configFile, "", "    ")
	require.NoError(t, err)
	require.NoError(t, ioutil.WriteFile(testingGenesisFilePath+jsonPosfix, genesisFileBz, 0777))
	require.NoError(t, ioutil.WriteFile(testingConfigFilePath+strconv.Itoa(n)+jsonPosfix, p2pFileBz, 0777))
}

func createConfigs(t *testing.T, numValidators int) (configs []modules.Config, genesisState modules.GenesisState) {
	configs = make([]modules.Config, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys[:], keys[:numValidators])
	genesisState = createGenesisState(t, valKeys)
	for i := range configs {
		configs[i] = modules.Config{
			Base: &modules.BaseConfig{
				RootDirectory: "",
				PrivateKey:    valKeys[i].String(),
			},
			P2P: &typesP2P.P2PConfig{
				PrivateKey:            valKeys[i].String(),
				ConsensusPort:         8080,
				UseRainTree:           true,
				IsEmptyConnectionType: true,
			},
		}
	}
	return
}

func createGenesisState(t *testing.T, valKeys []cryptoPocket.PrivateKey) modules.GenesisState {
	validators := make([]modules.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		val := &test_artifacts.MockActor{
			Address:         addr,
			PublicKey:       valKey.PublicKey().String(),
			GenericParam:    validatorId(t, i+1),
			StakedAmount:    "1000000000000000",
			PausedHeight:    0,
			UnstakingHeight: 0,
			Output:          addr,
		}
		validators[i] = val
	}
	return modules.GenesisState{
		PersistenceGenesisState: &test_artifacts.MockPersistenceGenesisState{
			Validators: validators,
		},
	}
}