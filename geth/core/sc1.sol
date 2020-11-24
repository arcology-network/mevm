pragma solidity ^0.5.0;

contract TestCase1
{
    int32 counter = 0;
    
    function Increase() public returns (int32){
        counter++;
        return counter;
    }
    
}