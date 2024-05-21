

#define ENTRIES_PER_SHARD 250000

enum CfaType {
  CFA_TYPE_RBP_OFFSET,
  CFA_TYPE_RSP_OFFSET,
  CFA_TYPE_EXPRESSION,
  CFA_TYPE_UNSUPPORTED,
};
typedef uint8_t CfaType;

enum RegType {
  REG_TYPE_UNDEFINED,
  REG_TYPE_SAME_VALUE,
  REG_TYPE_OFFSET,
  REG_TYPE_UNSUPPORTED,
};
typedef uint8_t RegType;

typedef struct ShardInfo {
  int32_t id;
  uint64_t pc_min;
  uint64_t pc_max;
} ShardInfo;

typedef struct ShardInfoList {
  struct ShardInfo info[40];
} ShardInfoList;

typedef struct UnwindEntry {
  uint64_t pc;
  CfaType cfa_type;
  RegType rbp_type;
  int16_t cfa_offset;
  int16_t rbp_offset;
} UnwindEntry;

typedef struct UnwindEntryShard {
  uint32_t len;
  struct UnwindEntry entries[ENTRIES_PER_SHARD];
} UnwindEntryShard;

void do_things(struct ShardInfoList, struct UnwindEntryShard);

void merge_stacks(int8_t *trace_str, size_t len, const int8_t *i_trace, const int8_t *u_trace);
