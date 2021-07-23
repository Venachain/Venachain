pragma experimental ABIEncoderV2;
pragma solidity 0.5.2;

contract PToken {

    struct G1Point {
        bytes32 x;
        bytes32 y;
    }
    
    uint256 public epochLength; 
    address public token;

    uint256 constant MAX = 4294967295; // 2^32 - 1 
    uint256 constant GROUP_ORDER = 0x30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001;
    uint256 constant FIELD_ORDER = 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47;
    
    mapping(bytes32 => G1Point[2]) acc; // main account mapping
    mapping(bytes32 => G1Point[2]) pending; // storage for pending transfers
    mapping(bytes32 => uint256) lastRollOver;
    bytes32[] nonceSet;
    G1Point[] pkSet;
    G1Point[] decoy;
    uint256 public lastGlobalUpdate = 0;
    uint256 public pkAmount = 0;
    uint256 private _totalSupply;

    
    constructor(uint256 _epochLength, address _token) public {
        epochLength = _epochLength;
        token = _token;
    }
    
    function totalSupply() public view returns (uint256) {
        return _totalSupply;
    }
    
    function transferfrom(uint256 from, uint256 to, uint256 value) internal {
        bytes memory header = bytes('{"func_name": "transfer", "func_params": ["uint64(');
        bytes memory param1 = uintToBytes(value);
        bytes memory middle = bytes(')", "string(');
        bytes memory param2 = uintToHexstr(from);
        bytes memory param3 = uintToHexstr(to);
        bytes memory tail = bytes(')"]}');
        
        bytes memory payload = new bytes(header.length + param1.length + middle.length + middle.length + 80 + tail.length);
        //uint paramLen = header.length + param1.length + middle.length + middle.length + 80 + tail.length;
        
        uint k = 0;
        uint i = 0;
        for (i = 0; i < header.length; i++){
            payload[k++] = header[i];
        }
        for (i = 0; i < param1.length; i++) {
            payload[k++] = param1[i];
        }
        for (i = 0; i < middle.length; i++) {
            payload[k++] = middle[i];
        }
        if (param2.length < 40) {
            for(uint pad = (40 - param2.length); pad > 0; pad--){
               payload[k++] = "0";
            }
        }
        for (i = 0; i < param2.length; i++) {
            payload[k++] = param2[i];
        }
        for (i = 0; i < middle.length; i++) {
            payload[k++] = middle[i];
        }
        if (param3.length < 40) {
            for(uint pad = (40 - param3.length); pad > 0; pad--){
               payload[k++] = "0";
            }
        }
        for (i = 0; i < param3.length; i++) {
            payload[k++] = param3[i];
        }
        for (i = 0; i < tail.length; i++) {
            payload[k++] = tail[i];
        }
        
        (bool success, bytes memory returnData) = address(token).call(payload);
        
        //require(success && keccak256(abi.encode(string(returnData))) != keccak256(abi.encode("0")));
        // uint[1] memory ret;
        // assembly{    
        //     if iszero(staticcall(gas, 0x0d, add(payload, 0x20), paramLen, ret, 0x20)) {
        //      revert(0, 0)
        //     }
        // }
        require(success && uint8(returnData[31]) == 0, "Bussiness contract transfer failed");
        
    }

    function withdrawVerify(G1Point memory c1, G1Point memory c2, G1Point memory y, G1Point memory u, bytes memory proof) internal view returns(bool) {
        uint epoch = lastGlobalUpdate;
        bytes memory e = new bytes(32);
        assembly{
            mstore(add(e, 32), epoch)
        }
        
        uint sender = uint256(msg.sender);
        bytes memory s = new bytes(32);
        assembly{
            mstore(add(s, 32), sender)
        }
        uint[1] memory ret;
        uint paramsLen = 256 + proof.length + 64;
        bytes memory params = new bytes(paramsLen);
        assembly{
            mstore(add(params, 0x20), mload(c1))
            mstore(add(params, 0x40), mload(add(c1, 0x20)))
            mstore(add(params, 0x60), mload(c2))
            mstore(add(params, 0x80), mload(add(c2, 0x20)))
            mstore(add(params, 0xa0), mload(y))
            mstore(add(params, 0xc0), mload(add(y, 0x20)))
            mstore(add(params, 0xe0), mload(u))
            mstore(add(params, 0x100), mload(add(u, 0x20)))
        }
        uint k = 256;
        uint i = 0;
        for(i = 0; i < 32; i++){
            params[k++] = e[i];
        }
        for(i = 0; i < 32; i++){
            params[k++] = s[i];
        }
        for(i = 0; i < proof.length; i++)
        {
            params[k++] = proof[i];
        }
        
        assembly{    
            if iszero(staticcall(gas, 0x0a, add(params, 0x20), paramsLen, ret, 0x20)) {
             revert(0, 0)
            }
        }
        return ret[0] != 0;
    }
    
    function transferVerify(G1Point[] memory C, G1Point memory D, G1Point[] memory y, G1Point memory u, bytes memory proof) internal view returns(bool){
        uint size = C.length + y.length + 2;
        uint yLen = y.length;
        uint paramsLen = proof.length + 64 + size * 0x40;
        bytes memory params = new bytes(paramsLen);
        assembly{
            mstore(add(params, 0x20), yLen)
        }
        uint k = C.length;
        for(uint i = 0; i < C.length; i++){
            G1Point memory tmp = C[i];
            bytes memory scratch = new bytes(64);
            assembly{
                mstore(add(scratch, 0x20), mload(tmp))
                mstore(add(scratch, 0x40), mload(add(tmp, 0x20)))
            }
            for(uint j = 0; j < 64; j++){
                params[i * 64 + j + 32] = scratch[j];
            }
        }
        for(uint i = 0; i < y.length; i++){
            G1Point memory tmp = y[i];
            bytes memory scratch = new bytes(64);
            assembly{
                mstore(add(scratch, 0x20), mload(tmp))
                mstore(add(scratch, 0x40), mload(add(tmp, 0x20)))
            }
            for(uint j = 0; j < 64; j++){
                params[(i + k) * 64 + j + 32] = scratch[j];
            }
        }
        k = C.length + y.length;
        bytes memory Dtmp = new bytes(64);
        assembly{
            mstore(add(Dtmp, 0x20), mload(D))
            mstore(add(Dtmp, 0x40), mload(add(D, 0x20)))
        }
        for(uint j = 0; j < 64; j++){
            params[k * 64 + j + 32] = Dtmp[j];
        }
        
        k++;
        bytes memory utmp = new bytes(64);
        assembly{
            mstore(add(utmp, 0x20), mload(u))
            mstore(add(utmp, 0x40), mload(add(u, 0x20)))
        }
        for(uint j = 0; j < 64; j++){
            params[k * 64 + j + 32] = utmp[j];
        }
        
        uint epoch = lastGlobalUpdate;
        bytes memory e = new bytes(32);
        assembly{
            mstore(add(e, 0x20), epoch)
        }
        for(uint i = 0; i < e.length; i++){
            params[i + 32 + size * 64] = e[i];
        }
        for(uint i = 0; i < proof.length; i++){
            params[i + 64 + size * 64] = proof[i];
        }
        uint[1] memory ret;
        assembly{    
            if iszero(staticcall(gas, 0x0c, add(params, 0x20), paramsLen, ret, 0x20)) {
             revert(0, 0)
            }
        }
        return ret[0] != 0;
    }
    
    function neg(uint256 x) internal pure returns (uint256) {
        return GROUP_ORDER - x;
    }
    function negEC(G1Point memory p) internal pure returns (G1Point memory) {
        return G1Point(p.x, bytes32(FIELD_ORDER - uint256(p.y))); 
    }
    
    function addEC(G1Point memory a, G1Point memory b) internal view returns(G1Point memory r) {
        assembly {
            let m := mload(0x40)
            mstore(m, mload(a))
            mstore(add(m, 0x20), mload(add(a, 0x20)))
            mstore(add(m, 0x40), mload(b))
            mstore(add(m, 0x60), mload(add(b, 0x20)))
            if iszero(staticcall(gas, 0x06, m, 0x80, r, 0x40)) {
                revert(0, 0)
            }
            
        }
    }
    
    function scalarmulEC(G1Point memory a, uint256 s) internal view returns(G1Point memory r) {
        assembly {
            let m := mload(0x40)
            mstore(m, mload(a))
            mstore(add(m, 0x20), mload(add(a, 0x20)))
            mstore(add(m, 0x40), s)
            if iszero(staticcall(gas, 0x07, m, 0x60, r, 0x40)) {
                revert(0, 0)
            }
        }
    }
    
    function eq(G1Point memory p1, G1Point memory p2) internal pure returns (bool) {
        return p1.x == p2.x && p1.y == p2.y;
    }
    
    function g() internal pure returns (G1Point memory) {
        return G1Point(0x14bcc435f49d130d189737f9762feb25c44ef5b886bef833e31a702af6be4748, 0x10cd33954522ad058f00a2553fd4e10d859fe125997e98adba777910dddc5322);
    }
    
    function verifySignature(G1Point memory pk, uint256 r, uint256 s) internal view returns (bool){
        // G1Point memory rpk = scalarmulEC(pk, r);
        // G1Point memory nrpk = negEC(rpk);
        // G1Point memory sG = scalarmulEC(g(), s);
        // G1Point memory K = addEC(sG, nrpk);
        // uint256 chanllenge = uint256(keccak256(abi.encode(address(this), K, pk)));
        // chanllenge = chanllenge % GROUP_ORDER;
        // emit Schnorr(K, chanllenge, abi.encode(address(this), K, pk));
        // if (chanllenge == r) {
        //     return true;
        // } 
        // return false;
        uint256 a = uint256(address(this));
        
        uint[1] memory ret;
        assembly{
            let m := mload(0x40)
            mstore(m, mload(pk))
            mstore(add(m, 0x20), mload(add(pk, 0x20)))
            mstore(add(m, 0x40), a)
            mstore(add(m,0x60), r)
            mstore(add(m,0x80), s)
            if iszero(staticcall(gas, 0x0b, m, 0xa0, ret, 0x20)) {
                revert(0, 0)
            }
        }
        return ret[0] != 0;    
        
    }
    

    
    function uintToBytes(uint v) internal pure returns (bytes memory ret) {
        if (v == 0) {
            ret = '0';
        }
        else {
            uint tmp = v;
            uint length;
            while (tmp != 0) {
                length++;
                tmp /= 10;
            }
            ret = new bytes(length);
            bytes32 strtmp;
            while (v > 0) {
                strtmp = bytes32(uint(strtmp) / (2 ** 8));
                strtmp |= bytes32(((v % 10) + 48) * 2 ** (8 * 31));
                v /= 10;
            }
            for(uint i = 0; i < length; i++){
                ret[i] = strtmp[i];
            }
        }
        return ret;
    }
    
    function uintToHexstr(uint i) internal pure returns (bytes memory ret) {
        if (i == 0) return "0";
        uint j = i;
        uint length;
        while (j != 0) {
            length++;
            j = j >> 4;
        }
        uint mask = 15;
        ret = new bytes(length);
        uint k = length - 1;
        while (i != 0){
            uint curr = (i & mask);
            bytes32 tmp = curr > 9 ? bytes32(55 + curr) : bytes32(48 + curr);
            ret[k--] = tmp[31];
            i = i >> 4;
        }
        return ret;
}

    function rollOver(bytes32 yHash) internal {
        uint256 e = block.number / epochLength;
        if (lastRollOver[yHash] < e) {
            G1Point[2][2] memory scratch = [acc[yHash], pending[yHash]];
            acc[yHash][0] = addEC(scratch[0][0], scratch[1][0]);
            acc[yHash][1] = addEC(scratch[0][1], scratch[1][1]);
            delete pending[yHash]; 
            lastRollOver[yHash] = e;
        }
        if (lastGlobalUpdate < e) {
            lastGlobalUpdate = e;
            delete nonceSet;
        }
    }

    function registered(bytes32 yHash) internal view returns (bool) {
        G1Point memory zero = G1Point(0, 0);
        G1Point[2][2] memory scratch = [acc[yHash], pending[yHash]];
        return !(eq(scratch[0][0], zero) && eq(scratch[0][1], zero) && eq(scratch[1][0], zero) && eq(scratch[1][1], zero));
    }
    
    function simulateAccounts(G1Point[] memory y, uint256 epoch) view public returns (G1Point[2][] memory accounts) {
        uint256 size = y.length;
        accounts = new G1Point[2][](size);
        for (uint256 i = 0; i < size; i++) {
            bytes32 yHash = keccak256(abi.encode(y[i]));
            accounts[i] = acc[yHash];
            if (lastRollOver[yHash] < epoch) {
                G1Point[2] memory scratch = pending[yHash];
                accounts[i][0] = addEC(accounts[i][0], scratch[0]);
                accounts[i][1] = addEC(accounts[i][1], scratch[1]);
            }
        }
    }
    
    // function getPublicKey(uint256[] memory index) view public returns(G1Point[] memory key){
    //     uint256 size = pkSet.length;
    //     uint256 pkNum = index.length;
    //     G1Point[] memory key;
    //     for(uint256 i = 0; i < pkNum; i++){
    //         uint256 j = index[i] % size;
    //         bytes32 yHash = keccak256(abi.encode(pkSet[j]));
    //         key[i] = pkSet[j];
    //         }
            
    //     }
    //     return key;
    // }
    function getPublicKey(uint256 index) view public returns(G1Point memory key){
        uint256 size = pkSet.length;
        uint256 i = index % size;
        return pkSet[i];
    }

    function register(G1Point memory y, uint256 r, uint256 s) public {
        bool result = verifySignature(y, r, s);
        require(result == true, "Invalid registration signature!");
        bytes32 yHash = keccak256(abi.encode(y));
        require(!registered(yHash), "Account already registered!");
        pending[yHash][0] = y;
        pending[yHash][1] = g();
        pkSet.push(y);
        pkAmount++;
    }

    function fund(G1Point memory y, uint256 bTransfer) public {
        bytes32 yHash = keccak256(abi.encode(y));
        require(registered(yHash), "Account not yet registered.");
        rollOver(yHash);

        require(bTransfer <= MAX, "Deposit amount out of range."); // uint, so other way not necessary?

        G1Point memory scratch = pending[yHash][0];
        G1Point memory tmp = scalarmulEC(g(), bTransfer);
        scratch = addEC(scratch, tmp);
        pending[yHash][0] = scratch;
        _totalSupply = _totalSupply + bTransfer;
        uint256 sender = uint256(msg.sender);
        uint256 receiver = uint256(address(this));
        transferfrom(sender, receiver, bTransfer);
        require(_totalSupply<= MAX, "Fund pushes contract past maximum value.");
    }

    function transfer(G1Point[] memory C, G1Point memory D, G1Point[] memory y, G1Point memory u, bytes memory proof) public {
        uint256 size = y.length * 3;
        G1Point[] memory Cipher = new G1Point[](size);
        require(C.length == y.length, "Input array length mismatch!");
        for (uint256 i = 0; i < y.length; i++) {
            bytes32 yHash = keccak256(abi.encode(y[i]));
            require(registered(yHash), "Account not yet registered.");
            rollOver(yHash);
            G1Point[2] memory scratch = pending[yHash]; 
            pending[yHash][0] = addEC(scratch[0], C[i]);
            pending[yHash][1] = addEC(scratch[1], D);
            scratch = acc[yHash];
            Cipher[i] = addEC(scratch[0], C[i]);
            Cipher[i + y.length] = addEC(scratch[1], D);
        }
        uint256 k = y.length * 2;
        for (uint256 i = 0; i < y.length; i++) {
            Cipher[k++] = C[i];
        }
        bytes32 uHash = keccak256(abi.encode(u));
        for (uint256 i = 0; i < nonceSet.length; i++) {
            require(nonceSet[i] != uHash, "Nonce already seen!");
        }
        nonceSet.push(uHash);
        bool result = transferVerify(Cipher, D, y, u, proof);
        require(result, "Transfer proof verification failed!");
    }

    function withdraw(G1Point memory y, uint256 bTransfer, G1Point memory u, bytes memory proof) public {
        bytes32 yHash = keccak256(abi.encode(y));
        require(registered(yHash), "Account not yet registered.");
        rollOver(yHash);

        require(0 <= bTransfer && bTransfer <= MAX, "Transfer amount out of range.");
        G1Point[2] memory scratch = pending[yHash];
        G1Point memory tmp = scalarmulEC(g(), neg(bTransfer));
        pending[yHash][0] = addEC(scratch[0], tmp);

        scratch = acc[yHash]; 
        scratch[0] = addEC(scratch[0], tmp);
        bytes32 uHash = keccak256(abi.encode(u));
        for (uint256 i = 0; i < nonceSet.length; i++) {
            require(nonceSet[i] != uHash, "Nonce already seen!");
        }
        nonceSet.push(uHash);
        
        G1Point memory c1 = scratch[0];
        G1Point memory c2 = scratch[1];
        bool result = withdrawVerify(c1, c2, y, u, proof);
        require(result, "Withdraw proof verification failed!");
        
        
        uint256 sender = uint256(address(this));
        uint256 receiver = uint256(msg.sender);
        transferfrom(sender, receiver, bTransfer);
        _totalSupply = _totalSupply - bTransfer;
    }

    
}
