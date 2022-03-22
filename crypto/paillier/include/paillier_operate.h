/*
 * @system:
 * @company: 万向区块链有限公司
 * @file:
 * @brief:
 * @author: jyz
 * @Date: 2022-03-15 18:06:08
 * @history:
 */
#ifndef __PAILLIER_OPERATE_H__

#define __PAILLIER_OPERATE_H__
#include "./include/paillier.h"

char *sPaillierWeightAdd(char **cipher, unsigned long int *weight, int length, char *pubKey);
char *sPaillierAdd(char **cipher, int length, char *pPubKey);
char *sPaillierMul(char *cipher, unsigned long int scalar, char *pPubKey);

char *paillierCipherToHex(paillier_ciphertext_t *ct);

paillier_ciphertext_t *paillierCipherFromHex(char *ct);
#endif