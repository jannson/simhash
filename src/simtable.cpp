#include "common.h"
#include "simtable.h"
#include "simhash.h"

#include <iostream>

using namespace std;
using namespace Simhash;

#define TB(a) ((Table*)(a))
#define VE(a) ((vector<unsigned long>*)a)

#ifdef __cplusplus
extern "C" {
#endif

SimTable SimTableInit(long d, LongVector lv) {
    Table* tb = new Table(d, *VE(lv));
    return (SimTable)tb;
}

void SimTableRelease(SimTable st) {
    if (NULL != st) {
        delete TB(st);
    }
}

void SimTableInsert(SimTable st, unsigned long hash) {
    TB(st)->insert(hash);
}

void SimTableInsertBulk(SimTable st, unsigned long *phash, long size) {
    TB(st)->insert(phash, size);
}

void SimTableRemove(SimTable st, unsigned long hash) {
    TB(st)->remove(hash);
}

unsigned long SimTableFind(SimTable st, unsigned long query) {
    return TB(st)->find(query);
}

void SimTableFindm(SimTable st, unsigned long query, LongVector lv) {
    return TB(st)->find(query, *VE(lv));
}

unsigned long SimTablePermute(SimTable st, unsigned long hash){
    return TB(st)->permute(hash);
}

unsigned long SimTableUnpermute(SimTable st, unsigned long hash){
    return TB(st)->unpermute(hash);
}

unsigned long SimTableSearchMask(SimTable st) {
    return TB(st)->get_search_mask();
}

LongVector LongVectorInit() {
    return new vector<unsigned long>();
}

void LongVectorReserve(LongVector lv, long s) {
    VE(lv)->reserve(s);
}

void LongVectorAdd(LongVector lv, unsigned long v) {
    VE(lv)->push_back(v);
}

void LongVectorSet(LongVector lv, int i, unsigned long v) {
    (*VE(lv))[i] = v;
}

unsigned long LongVectorGet(LongVector lv, int i) {
    return (*VE(lv))[i];
}

void LongVectorRelease(LongVector lv) {
    if (NULL != lv) {
        delete VE(lv);
    }
}

long LongVectorLen(LongVector lv) {
    return VE(lv)->size();
}

unsigned long* LongVector2Array(LongVector lv, long *inLen) {
    if (NULL != inLen) {
        *inLen = VE(lv)->size();
    }
    return &((*VE(lv))[0]);
}

#ifdef __cplusplus
}
#endif

