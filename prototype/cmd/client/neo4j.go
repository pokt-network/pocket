package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/mindstand/gogm/v2"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"io/ioutil"
	"log"
	"net"
	consensus_types "pocket/consensus/types"
	p2p_types "pocket/p2p/pre_p2p/types"
)

func getNodeData(network p2p_types.Network) (node_data []*consensus_types.ConsensusNodeState) {
	for _, networkModule := range network.GetAddrBook() {
		conn, err := net.DialTCP("tcp", nil, networkModule.DebugAddr)
		if err != nil {
			log.Println("Error connecting to peer debug port: ", err)
			continue
		}
		defer conn.Close()

		data, err := ioutil.ReadAll(conn)
		if err != nil {
			log.Println("Error reading from conn: ", err)
			return
		}

		var buff = bytes.NewBuffer(data)
		dec := gob.NewDecoder(buff)
		consensusNodeState := consensus_types.ConsensusNodeState{}
		if err = dec.Decode(&consensusNodeState); err != nil {
			log.Println("[ERROR] Error decoding: ", err)
		}

		node_data = append(node_data, &consensusNodeState)
	}
	return
}

func dropAll() {
	driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("root", "", ""), func(c *neo4j.Config) { c.Encrypted = false })
	if err != nil {
		log.Panicln(err)
	}
	defer driver.Close()

	session, err := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	if err != nil {
		log.Panicln(err)
	}
	defer session.Close()

	// MATCH (n:ConsensusNodeState) RETURN n LIMIT 100
	res, err := session.Run("MATCH (n:ConsensusNodeState) DETACH DELETE n", map[string]interface{}{})
	if err != nil {
		log.Panicln(err)
	}
	log.Println(res)
}

// See https://github.com/mindstand/gogm?ref=golangrepo.com as a reference
func DumpToNeo4j(network p2p_types.Network) {
	config := gogm.Config{
		Host:     "localhost",
		Port:     7687,
		Protocol: "bolt", // {neo4j neo4j+s, neo4j+ssc, bolt, bolt+s and bolt+ssc}
		Username: "neo4j",
		// Password:           "",
		PoolSize: 50,
		// IndexStrategy:      gogm.VALIDATE_INDEX, // {VALIDATE_INDEX, ASSERT_INDEX, IGNORE_INDEX}
		IndexStrategy:      gogm.IGNORE_INDEX, // {VALIDATE_INDEX, ASSERT_INDEX, IGNORE_INDEX}
		TargetDbs:          nil,
		Logger:             gogm.GetDefaultLogger(),
		LogLevel:           "DEBUG",
		EnableDriverLogs:   false,
		EnableLogParams:    false,
		OpentracingEnabled: false,
	}

	_gogm, err := gogm.New(&config, gogm.DefaultPrimaryKeyStrategy, &consensus_types.ConsensusNodeState{})
	if err != nil {
		panic(err)
	}
	sess, err := _gogm.NewSessionV2(gogm.SessionConfig{AccessMode: gogm.AccessModeWrite})
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	nodeData := getNodeData(network)

	// Drop all existing nodes.
	dropAll()

	// Save all the nodes.
	for _, consensusNodeState := range nodeData {
		if err = sess.Save(context.Background(), consensusNodeState); err != nil {
			panic(err)
		}
	}
}
