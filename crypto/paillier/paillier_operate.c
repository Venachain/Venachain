/*
 * @system:
 * @company: 万向区块链有限公司
 * @file:
 * @brief:
 * @author: jyz
 * @Date: 2022-03-15 14:57:23
 * @history:
 */
#include "./include/log.h"
#include "./include/paillier_operate.h"
#include <stdlib.h>

char *sPaillierWeightAdd(char **cipher, unsigned long int *weight, int length, char *pPubKey)
{
    // encrypted 0
    paillier_ciphertext_t *sum = paillier_create_enc_zero();

    paillier_pubkey_t *pubKey = paillier_pubkey_from_hex(pPubKey);
    paillier_ciphertext_t *mCipher;
    paillier_plaintext_t *mWeight;

    for (int i = 0; i < length; i++)
    {
        if (*cipher == NULL)
        {
            return NULL;
        }
        mCipher = paillierCipherFromHex(*cipher++);
        mWeight = paillier_plaintext_from_ui(*weight++);

        if (mpz_cmp(mCipher->c, pubKey->n_squared) >= 0)
        {
            log_warn("cipher is greater than or equal to n_squared");
            return NULL;
        }
        // mCipher * mWeight
        paillier_exp(pubKey, mCipher, mCipher, mWeight);
        // Sum the encrypted values by multiplying the ciphertexts
        paillier_mul(pubKey, sum, sum, mCipher);
    }
    char *res = paillierCipherToHex(sum);

    // clean
    paillier_freeciphertext(sum);
    paillier_freeciphertext(mCipher);
    paillier_freeplaintext(mWeight);
    paillier_freepubkey(pubKey);

    return res;
}

char *sPaillierAdd(char **cipher, int length, char *pPubKey)
{
    // encrypted 0
    paillier_ciphertext_t *sum = paillier_create_enc_zero();

    paillier_pubkey_t *pubKey = paillier_pubkey_from_hex(pPubKey);
    paillier_ciphertext_t *mCipher;

    for (int i = 0; i < length; i++)
    {
        mCipher = paillierCipherFromHex(*cipher++);
        if (mpz_cmp(mCipher->c, pubKey->n_squared) >= 0)
        {

            log_warn("cipher is greater than or equal to n_squared");
            return NULL;
        }
        // Sum the encrypted values by multiplying the ciphertexts
        paillier_mul(pubKey, sum, sum, mCipher);
    }
   char *res = paillierCipherToHex(sum);

    // clean
    paillier_freeciphertext(sum);
    paillier_freeciphertext(mCipher);
    paillier_freepubkey(pubKey);

    return res;
}

char *sPaillierMul( char *cipher, unsigned long int scalar, char *pPubKey)
{
    // encrypted 0
    paillier_ciphertext_t *sum = paillier_create_enc_zero();

    paillier_plaintext_t *paillierPlainTextScalar = paillier_plaintext_from_ui(scalar);
    paillier_pubkey_t *pubKey = paillier_pubkey_from_hex(pPubKey);
    paillier_ciphertext_t *mCipher = paillierCipherFromHex(cipher);
    if (mpz_cmp(mCipher->c, pubKey->n_squared) >= 0)
    {
        return NULL;
        log_warn("cipher is greater than or equal to n_squared");
    }

    paillier_exp(pubKey, sum, mCipher, paillierPlainTextScalar);

    char *res = paillierCipherToHex(sum);

    // clean
    paillier_freeciphertext(sum);
    paillier_freeciphertext(mCipher);
    paillier_freeplaintext(paillierPlainTextScalar);
    paillier_freepubkey(pubKey);

    return res;
}

char *paillierCipherToHex(paillier_ciphertext_t *ct)
{
    return mpz_get_str(0, 16, ct->c);
}

paillier_ciphertext_t *paillierCipherFromHex(char *ct)
{
    paillier_ciphertext_t *pt = (paillier_ciphertext_t *)malloc(sizeof(paillier_ciphertext_t));
    mpz_init(pt->c);
    mpz_set_str(pt->c, ct, 16);
    return pt;
}
