#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>
#define RETURN_CHARARRAY(src, size)       \
    do                                    \
    {                                     \
        char *buf = (char *)malloc(size); \
        memset(buf, 0, size);             \
        strcpy(buf, src);                 \
        return buf;                       \
    } while (0)

namespace demo {
char mapName[] = "UserState";

//以address为key，balance为value
typedef bcwasm::db::Map<mapName, std::string, uint64_t> userMap_t;
class BcToken : public bcwasm::Contract
{
private:
    userMap_t userMap;
    uint64_t initbalance = 4294967295;
public:
    BcToken() {}
    /// 实现父类: bcwasm::Contract 的虚函数
    /// 该函数在合约首次发布时执行，仅调用一次
    void init()
    {
        std::string addr = bcwasm::origin().toString();
        formatAddress(addr);

        bcwasm::setState("owner", addr);
        userMap.insert(addr, initbalance);
        bcwasm::println(addr);
        bcwasm::println("Token init success...");
    }
    //event
    BCWASM_EVENT(Register, const char *)
    BCWASM_EVENT(Transfer, const char *)
    enum Code
    {
        SUCCESS,
        FAILURE,
        HAS_REGISTERED,
        NOT_REGISTERED
    };

public:
    int64_t userregister(uint64_t balance)
    {
        std::string addr = bcwasm::origin().toString();
        formatAddress(addr);
        //check register information
        if (isRegistered(addr.c_str()))
        {
            BCWASM_EMIT_EVENT(Register, "Record exist");
            return HAS_REGISTERED;
        }
        userMap.insert(addr, balance);
        BCWASM_EMIT_EVENT(Register, "Register succeed!");
        return SUCCESS;
    }

    int64_t transfer(uint64_t value, const char *sender, const char *receiver)
    {
        bcwasm::println("sender:");
        bcwasm::println(sender);
        bcwasm::println("receiver:");
        bcwasm::println(receiver);
        std::string s = std::string(sender);
        std::string r = std::string(receiver);
        formatAddress(s);
        formatAddress(r);
        bcwasm::println(s);
        bcwasm::println(r);
        if (!isRegistered(s.c_str()))
        {
            BCWASM_EMIT_EVENT(Transfer, "Record not exist");
            return FAILURE;
        }
        
        const uint64_t *SenderPtr = userMap.find(s);
        
        if (*SenderPtr >= value)
        {  
            uint64_t balance = *SenderPtr - value;
            userMap.update(s, balance);
            if(!isRegistered(r.c_str()))
            {
                registerReceiver(value, r);
            } else {
                const uint64_t *ReceiverPtr = userMap.find(r);
                uint64_t tmp = *ReceiverPtr + value;
                userMap.update(r, tmp);
            }
            return SUCCESS;
        }
        return FAILURE;
    }

    uint64_t getBalance(const char *addr) const
    {
        uint64_t balance = 0;
        if (!isRegistered(addr))
        {
            return NULL;
        }
        const uint64_t *UserPtr = userMap.find(addr);
        balance = *UserPtr;
        return balance;
    }

    int checkBalance(const char *addr, uint64_t value) const
    {
        int result = FAILURE;
        if (!isRegistered(addr))
        {
            return result;
        }
        const uint64_t *InfoPtr = userMap.find(addr);
        if (*InfoPtr >= value){
            result = SUCCESS;
        }
        return result;
    }




private:
    void formatAddress(std::string& addr) {
        if (addr.find("0x") != 0)
            addr = std::string("0x") + addr;
        std::transform(addr.begin(), addr.end(), addr.begin(), ::tolower);
    }
    //获取合约调用者
    std::string getOrigin()
    {
        std::string origin = bcwasm::origin().toString();
        formatAddress(origin);
        return origin;
    }

    bool isRegistered(const char *addr) const
    {
        const uint64_t *InfoPtr = userMap.find(addr);
        if (nullptr != InfoPtr)
        {
            return true;
        }
        else
        {
            return false;
        }
    }

    int registerReceiver(uint64_t balance, std::string addr)
    {

        //check register information
        if (isRegistered(addr.c_str()))
        {
            BCWASM_EMIT_EVENT(Register, "Record exist");
            return HAS_REGISTERED;
        }
        userMap.insert(addr, balance);
        BCWASM_EMIT_EVENT(Register, "Register succeed!");
        return SUCCESS;
    }

};
} // namespace demo
// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(demo::BcToken, userregister)
BCWASM_ABI(demo::BcToken, transfer)
BCWASM_ABI(demo::BcToken, checkBalance)
BCWASM_ABI(demo::BcToken, getBalance)

//bcwasm autogen begin
extern "C" { 
long long userregister(unsigned long long balance) {
demo::BcToken BcToken_bcwasm;
return BcToken_bcwasm.userregister(balance);
}
long long transfer(unsigned long long value,const char * sender,const char * receiver) {
demo::BcToken BcToken_bcwasm;
return BcToken_bcwasm.transfer(value,sender,receiver);
}
unsigned long long getBalance(const char * addr) {
demo::BcToken BcToken_bcwasm;
return BcToken_bcwasm.getBalance(addr);
}
int checkBalance(const char * addr,unsigned long long value) {
demo::BcToken BcToken_bcwasm;
return BcToken_bcwasm.checkBalance(addr,value);
}
void init() {
demo::BcToken BcToken_bcwasm;
BcToken_bcwasm.init();
}

}
//bcwasm autogen end