pragma solidity ^0.5.0;

// Declaration of system level API.
contract ConcurrentArray {
  function create(string calldata id) external;
  function append(string calldata id, int64 value) external;
  function set(string calldata id, int32 index, int64 value) external;
  function get(string calldata id, int32 index) external returns (int64);
}

contract client {
  // System reserved address.
  address apiAddr = address(0x80);
  
  event LogGet(int32, int64);
  
  constructor() public {
    // Bind the address.
    ConcurrentArray ca = ConcurrentArray(apiAddr);
    // Create a new concurrent array for use in the scope of 'client'.
    ca.create("users");
  }
  
  // This function can be called concurrently.
  function add(int64 initValue) public {
    ConcurrentArray ca = ConcurrentArray(apiAddr);
    ca.append("users", initValue);
  }

  // This function can be called concurrently.
  function set(int32 index, int64 value) public {
    ConcurrentArray ca = ConcurrentArray(apiAddr);
    ca.set("users", index, value);
  }

  // This function can be called concurrently.
  function get(int32 index) public returns (int64) {
    ConcurrentArray ca = ConcurrentArray(apiAddr);
    int64 value = ca.get("users", index);
    emit LogGet(index, value);
  }
}
