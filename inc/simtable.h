#ifndef __SIMTABLE_H_
#define __SIMTABLE_H_

#ifdef __cplusplus
extern "C" {
#endif

#define SimTable void*
#define LongVector void*

SimTable SimTableInit(long d, LongVector lv);
void SimTableRelease(SimTable st);
void SimTableInsert(SimTable st, unsigned long hash);
void SimTableInsertBulk(SimTable st, unsigned long *phash, long size);
void SimTableRemove(SimTable st, unsigned long hash);
unsigned long SimTableFind(SimTable st, unsigned long query);
void SimTableFindm(SimTable st, unsigned long query, LongVector lv);
unsigned long SimTablePermute(SimTable st, unsigned long hash);
unsigned long SimTableUnpermute(SimTable st, unsigned long hash);
unsigned long SimTableSearchMask(SimTable st);

LongVector LongVectorInit();
void LongVectorReserve(LongVector lv, long s);
void LongVectorAdd(LongVector lv, unsigned long v);
void LongVectorSet(LongVector lv, int i, unsigned long v);
unsigned long LongVectorGet(LongVector lv, int i);
unsigned long* LongVector2Array(LongVector lv, long *inLen);
void LongVectorRelease(LongVector lv);
long LongVectorLen(LongVector lv);

#ifdef __cplusplus
}
#endif

#endif

