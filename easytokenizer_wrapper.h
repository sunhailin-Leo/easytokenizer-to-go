//
// Created by sunhailin-Leo on 2023/3/6.
//
#include <stdlib.h>
#include <stdbool.h>
#ifdef __cplusplus
extern "C" {
#endif
typedef void *EasyTokenizer;

EasyTokenizer initTokenizer(char* vocab_path, bool do_lower_case, bool codepoint_level);

void encode(EasyTokenizer tokenizer, char* text, bool add_cls_sep, bool truncation, int max_length, int* output_data);

void encodeWithIds(
        EasyTokenizer tokenizer,
        char* text,
        int** input_ids,
        int* num_input_ids,
        int** token_type_ids,
        int* num_token_type_ids,
        int** attention_mask,
        int* num_attention_mask,
        int** offsets,
        int* num_offsets,
        bool add_cls_sep,
        bool padding,
        bool truncation,
        int max_length);

void wordPieceTokenize(EasyTokenizer tokenizer, char* text, char*** tokens, int* num_tokens, int** offsets, int* num_offsets);

#ifdef __cplusplus
}
#endif