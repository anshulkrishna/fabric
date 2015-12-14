/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package trie

import (
	"fmt"
	"github.com/openblockchain/obc-peer/openchain/ledger/statemgmt"
	"github.com/openblockchain/obc-peer/openchain/ledger/testutil"
	"testing"
)

func TestStateTrie_ComputeHash_AllInMemory_NoContents(t *testing.T) {
	testDBWrapper.CreateFreshDB(t)
	stateTrie := NewStateTrie()
	stateTrieTestWrapper := &stateTrieTestWrapper{stateTrie, t}
	hash := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(statemgmt.NewStateDelta())
	testutil.AssertEquals(t, hash, nil)
}

func TestStateTrie_ComputeHash_AllInMemory(t *testing.T) {
	testDBWrapper.CreateFreshDB(t)
	stateTrie := NewStateTrie()
	stateTrieTestWrapper := &stateTrieTestWrapper{stateTrie, t}
	stateDelta := statemgmt.NewStateDelta()

	// Test1 - Add a few keys
	stateDelta.Set("chaincodeID1", "key1", []byte("value1"))
	stateDelta.Set("chaincodeID1", "key2", []byte("value2"))
	stateDelta.Set("chaincodeID2", "key3", []byte("value3"))
	stateDelta.Set("chaincodeID2", "key4", []byte("value4"))
	rootHash1 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)

	hash1 := computeTestHash(newTrieKey("chaincodeID1", "key1").getEncodedBytes(), []byte("value1"))
	hash2 := computeTestHash(newTrieKey("chaincodeID1", "key2").getEncodedBytes(), []byte("value2"))
	hash3 := computeTestHash(newTrieKey("chaincodeID2", "key3").getEncodedBytes(), []byte("value3"))
	hash4 := computeTestHash(newTrieKey("chaincodeID2", "key4").getEncodedBytes(), []byte("value4"))

	hash1_hash2 := computeTestHash(hash1, hash2)
	hash3_hash4 := computeTestHash(hash3, hash4)
	expectedRootHash1 := computeTestHash(hash1_hash2, hash3_hash4)
	testutil.AssertEquals(t, rootHash1, expectedRootHash1)
	stateTrie.ClearWorkingSet()

	// Test2 - Add one more key
	fmt.Println("\n-- Add one more key exiting key --- ")
	stateDelta.Set("chaincodeID3", "key5", []byte("value5"))
	rootHash2 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	hash5 := computeTestHash(newTrieKey("chaincodeID3", "key5").getEncodedBytes(), []byte("value5"))
	expectedRootHash2 := computeTestHash(hash1_hash2, hash3_hash4, hash5)
	testutil.AssertEquals(t, rootHash2, expectedRootHash2)
	stateTrie.ClearWorkingSet()

	// Test3 - Remove one of the existing keys
	fmt.Println("\n-- Remove an exiting key --- ")
	stateDelta.Delete("chaincodeID2", "key4")
	rootHash3 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	expectedRootHash3 := computeTestHash(hash1_hash2, hash3, hash5)
	testutil.AssertEquals(t, rootHash3, expectedRootHash3)
	stateTrie.ClearWorkingSet()
}

func TestStateTrie_GetSet_WithDB(t *testing.T) {
	testDBWrapper.CreateFreshDB(t)
	stateTrie := NewStateTrie()
	stateTrieTestWrapper := &stateTrieTestWrapper{stateTrie, t}
	stateDelta := statemgmt.NewStateDelta()
	stateDelta.Set("chaincodeID1", "key1", []byte("value1"))
	stateDelta.Set("chaincodeID1", "key2", []byte("value2"))
	stateDelta.Set("chaincodeID2", "key3", []byte("value3"))
	stateDelta.Set("chaincodeID2", "key4", []byte("value4"))
	stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()
	testutil.AssertEquals(t, stateTrieTestWrapper.Get("chaincodeID1", "key1"), []byte("value1"))
}

func TestStateTrie_ComputeHash_WithDB_Spread_Keys(t *testing.T) {
	testDBWrapper.CreateFreshDB(t)
	stateTrie := NewStateTrie()
	stateTrieTestWrapper := &stateTrieTestWrapper{stateTrie, t}

	// Add a few keys and write to DB
	stateDelta := statemgmt.NewStateDelta()
	stateDelta.Set("chaincodeID1", "key1", []byte("value1"))
	stateDelta.Set("chaincodeID1", "key2", []byte("value2"))
	stateDelta.Set("chaincodeID2", "key3", []byte("value3"))
	stateDelta.Set("chaincodeID2", "key4", []byte("value4"))
	stateTrie.PrepareWorkingSet(stateDelta)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()

	/////////////////////////////////////////////////////////
	// Test1 - Add a non-existing key
	/////////////////////////////////////////////////////////
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Set("chaincodeID3", "key5", []byte("value5"))
	rootHash1 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	expected_hash1 := computeTestHash(newTrieKey("chaincodeID1", "key1").getEncodedBytes(), []byte("value1"))
	expected_hash2 := computeTestHash(newTrieKey("chaincodeID1", "key2").getEncodedBytes(), []byte("value2"))
	expected_hash3 := computeTestHash(newTrieKey("chaincodeID2", "key3").getEncodedBytes(), []byte("value3"))
	expected_hash4 := computeTestHash(newTrieKey("chaincodeID2", "key4").getEncodedBytes(), []byte("value4"))
	expected_hash1_hash2 := computeTestHash(expected_hash1, expected_hash2)
	expected_hash3_hash4 := computeTestHash(expected_hash3, expected_hash4)
	expected_hash5 := computeTestHash(newTrieKey("chaincodeID3", "key5").getEncodedBytes(), []byte("value5"))
	expectedRootHash1 := computeTestHash(expected_hash1_hash2, expected_hash3_hash4, expected_hash5)
	testutil.AssertEquals(t, rootHash1, expectedRootHash1)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()

	/////////////////////////////////////////////////////////
	// Test2 - Change value of an existing key
	/////////////////////////////////////////////////////////
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Set("chaincodeID2", "key4", []byte("value4-new"))
	rootHash2 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	expected_hash4 = computeTestHash(newTrieKey("chaincodeID2", "key4").getEncodedBytes(), []byte("value4-new"))
	expected_hash3_hash4 = computeTestHash(expected_hash3, expected_hash4)
	expectedRootHash2 := computeTestHash(expected_hash1_hash2, expected_hash3_hash4, expected_hash5)
	testutil.AssertEquals(t, rootHash2, expectedRootHash2)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()

	/////////////////////////////////////////////////////////
	// Test3 - Change value of another existing key
	/////////////////////////////////////////////////////////
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Set("chaincodeID1", "key1", []byte("value1-new"))
	rootHash3 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	expected_hash1 = computeTestHash(newTrieKey("chaincodeID1", "key1").getEncodedBytes(), []byte("value1-new"))
	expected_hash1_hash2 = computeTestHash(expected_hash1, expected_hash2)
	expectedRootHash3 := computeTestHash(expected_hash1_hash2, expected_hash3_hash4, expected_hash5)
	testutil.AssertEquals(t, rootHash3, expectedRootHash3)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()

	/////////////////////////////////////////////////////////
	// Test4 - Delete an existing existing key
	/////////////////////////////////////////////////////////
	fmt.Println("\n -- Delete an existing key --- ")
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Delete("chaincodeID3", "key5")
	rootHash4 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	expectedRootHash4 := computeTestHash(expected_hash1_hash2, expected_hash3_hash4)
	testutil.AssertEquals(t, rootHash4, expectedRootHash4)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()
	// Delete should remove the key from db because, this key has no value and no children
	testutil.AssertNil(t, testDBWrapper.GetFromStateCF(t, newTrieKey("chaincodeID3", "key5").getEncodedBytes()))

	/////////////////////////////////////////////////////////
	// Test5 - Delete another existing existing key
	/////////////////////////////////////////////////////////
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Delete("chaincodeID2", "key4")
	rootHash5 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	expectedRootHash5 := computeTestHash(expected_hash1_hash2, expected_hash3)
	testutil.AssertEquals(t, rootHash5, expectedRootHash5)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()
	testutil.AssertNil(t, testDBWrapper.GetFromStateCF(t, newTrieKey("chaincodeID2", "key4").getEncodedBytes()))
}

func TestStateTrie_ComputeHash_WithDB_Staggered_Keys(t *testing.T) {
	testDBWrapper.CreateFreshDB(t)
	stateTrie := NewStateTrie()
	stateTrieTestWrapper := &stateTrieTestWrapper{stateTrie, t}

	/////////////////////////////////////////////////////////
	// Test1 - Add a few staggered keys
	/////////////////////////////////////////////////////////
	stateDelta := statemgmt.NewStateDelta()
	stateDelta.Set("ID", "key1", []byte("value_key1"))
	stateDelta.Set("ID", "key", []byte("value_key"))
	stateDelta.Set("ID", "k", []byte("value_k"))
	expectedHash_key1 := computeTestHash(newTrieKey("ID", "key1").getEncodedBytes(), []byte("value_key1"))
	expectedHash_key := computeTestHash(newTrieKey("ID", "key").getEncodedBytes(), []byte("value_key"), expectedHash_key1)
	expectedHash_k := computeTestHash(newTrieKey("ID", "k").getEncodedBytes(), []byte("value_k"), expectedHash_key)
	rootHash1 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	testutil.AssertEquals(t, rootHash1, expectedHash_k)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()

	/////////////////////////////////////////////////////////
	// Test2 - Add a new key in path of existing staggered keys
	/////////////////////////////////////////////////////////
	fmt.Println("- Add a new key in path of existing staggered keys -")
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Set("ID", "ke", []byte("value_ke"))
	expectedHash_ke := computeTestHash(newTrieKey("ID", "ke").getEncodedBytes(), []byte("value_ke"), expectedHash_key)
	expectedHash_k = computeTestHash(newTrieKey("ID", "k").getEncodedBytes(), []byte("value_k"), expectedHash_ke)
	rootHash2 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	testutil.AssertEquals(t, rootHash2, expectedHash_k)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()

	/////////////////////////////////////////////////////////
	// Test3 - Change value of one of the existing keys
	/////////////////////////////////////////////////////////
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Set("ID", "ke", []byte("value_ke_new"))
	expectedHash_ke = computeTestHash(newTrieKey("ID", "ke").getEncodedBytes(), []byte("value_ke_new"), expectedHash_key)
	expectedHash_k = computeTestHash(newTrieKey("ID", "k").getEncodedBytes(), []byte("value_k"), expectedHash_ke)
	rootHash3 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	testutil.AssertEquals(t, rootHash3, expectedHash_k)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()

	/////////////////////////////////////////////////////////
	// Test4 - delete one of the existing keys
	/////////////////////////////////////////////////////////
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Delete("ID", "ke")
	expectedHash_k = computeTestHash(newTrieKey("ID", "k").getEncodedBytes(), []byte("value_k"), expectedHash_key)
	rootHash4 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	testutil.AssertEquals(t, rootHash4, expectedHash_k)
	stateTrieTestWrapper.PersistChangesAndResetInMemoryChanges()
	// Delete should not remove the key from db because, this key has children
	testutil.AssertNotNil(t, testDBWrapper.GetFromStateCF(t, newTrieKey("ID", "ke").getEncodedBytes()))
	testutil.AssertEquals(t, rootHash1, rootHash4)

	//////////////////////////////////////////////////////////////
	// Test4 -  Add one more key as a sibling of an intermediate node
	//////////////////////////////////////////////////////////////
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Set("ID", "kez", []byte("value_kez"))
	expectedHash_kez := computeTestHash(newTrieKey("ID", "kez").getEncodedBytes(), []byte("value_kez"))
	expectedHash_ke = computeTestHash(expectedHash_key, expectedHash_kez)
	expectedHash_k = computeTestHash(newTrieKey("ID", "k").getEncodedBytes(), []byte("value_k"), expectedHash_ke)
	rootHash5 := stateTrieTestWrapper.PrepareWorkingSetAndComputeCryptoHash(stateDelta)
	testutil.AssertEquals(t, rootHash5, expectedHash_k)
}