package cc

var hdr_u_h = `
typedef signed char schar;
typedef unsigned char uchar;
typedef unsigned short ushort;
typedef unsigned int uint;
typedef unsigned long ulong;
typedef long long vlong;
typedef unsigned long long uvlong;

typedef schar int8;
typedef uchar uint8;
typedef short int16;
typedef ushort uint16;
typedef long int32;
typedef ulong uint32;
typedef vlong int64;
typedef uvlong uint64;
typedef float float32;
typedef double float64;
typedef unsigned long uintptr;

typedef schar s8int;
typedef uchar u8int;
typedef short s16int;
typedef ushort u16int;
typedef long s32int;
typedef ulong u32int;
typedef vlong s64int;
typedef uvlong u64int;

void *nil;

typedef struct va_list *va_list;
`

var hdr_libc_h = `
#include <fmt.h>

int nelem(void*);

int memcmp(void*, void*, long);
void *memset(void*, int, long);
int strcmp(char*, char*);
int strncmp(char*, char*, int);
char *strcpy(char*, char*);
char *smprint(char*, ...);
void strcat(char*, char*);

int atoi(char*);
ulong strtoul(char*, char**, int);
uvlong strtoull(char*, char**, int);
char *getenv(char*);
int getwd(char*, int);
double cputime(void);

enum
{
	AEXIST = 0,
};

int errstr(char*, uint);
void werrstr(char*, ...);

void exits(char*);
void sysfatal(char*, ...);
char *strstr(char*, char*);
int strlen(char*);
void memmove(void*, void*, int);
char *strdup(char*);
void *malloc(int);
void *calloc(int, int);
void *realloc(void*, int);
void free(void*);

void va_start(va_list, void*);
void va_end(va_list);
void qsort(void *base, int nmemb, int size, int (*compar)(const void *, const void *));
char *GOEXPERIMENT;
void setfcr(int);
void notify(void*);
void signal(int, void*);
uintptr getcallerpc(void*);

enum
{
	OREAD,
	OWRITE,
	ORDWR,
	SIGBUS,
	SIGSEGV,
	NDFLT,
	FPPDBL,
	FPRNR,
	BOM,
	HEADER_IO,
};

extern	void	flagcount(char*, char*, int*);
extern	void	flagint32(char*, char*, int32*);
extern	void	flagint64(char*, char*, int64*);
extern	void	flagstr(char*, char*, char**);
extern	void	flagparse(int*, char***, void (*usage)(void));
extern	void	flagfn0(char*, char*, void(*fn)(void));
extern	void	flagfn1(char*, char*, void(*fn)(char*));
extern	void	flagfn2(char*, char*, void(*fn)(char*, char*));
extern	void	flagprint(int);
extern	char*	strecpy(char*, char*, char*);
extern	void	abort(void);
extern	int	remove(const char*);
extern	char*	getgoos(void);
extern	char*	getgoarch(void);
extern	char*	getgoroot(void);
extern	char*	getgoversion(void);
extern	char*	getgoarm(void);
extern	char*	getgo386(void);
extern	char*	getgoextlinkenabled(void);
extern	char*	strchr(char*, int);
extern	char*	strrchr(char*, int);
extern	double	floor(double);
extern	double	ldexp(double, int);
extern	double	frexp(double, int*);
extern	double	pow(double, double);

extern	int	access(char*, int);
extern	int	isdigit(int);
extern	int	isalpha(int);
extern	int	isalnum(int);
extern	int	getfields(char*, char**, int, int, char*);
extern	char*	cleanname(char*);
extern	int	noted(int);

`

var hdr_extra_go_h = `
extern Node *N;
extern Sym *S;
extern Type *T;
extern Label *L;
//extern Case *C;
extern Prog *P;

enum
{
	BITS = 5,
	NVAR = BITS*4*8,
};
`
