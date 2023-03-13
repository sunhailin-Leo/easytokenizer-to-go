//
// Created by sunhailin-Leo on 2023/3/6.
//
#include <string>
#include <vector>
#include "easytokenizer_wrapper.h"
#include "tokenizer/tokenizer.h"

EasyTokenizer initTokenizer(char* vocab_path, bool do_lower_case, bool codepoint_level) {
    tokenizer::Tokenizer *tk = new tokenizer::Tokenizer(vocab_path, do_lower_case, codepoint_level);
    return (void *) tk;
}

void encode(EasyTokenizer tokenizer, char* text, bool add_cls_sep, bool truncation, int max_length, int* output_data) {
    auto data = ((tokenizer::Tokenizer *) tokenizer)->encode(text, add_cls_sep, truncation, max_length);

    std::vector<int>* vec = new std::vector<int>(data.begin(), data.end());
    int size = vec->size();
    for (int i = 0; i < size; i++) {
        output_data[i] = (*vec)[i];
    }

    delete vec;
}

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
        int max_length) {
    std::vector<tokenizer::SizeT> input_ids_vector;
    std::vector<tokenizer::SizeT> token_type_ids_vector;
    std::vector<tokenizer::SizeT> attention_mask_vector;
    std::vector<tokenizer::SizeT> offsets_vector;

    ((tokenizer::Tokenizer *) tokenizer)->encodeWithIds(text, input_ids_vector, token_type_ids_vector, attention_mask_vector, offsets_vector, add_cls_sep, padding, truncation, max_length);

    // input_ids_vector
    *num_input_ids = input_ids_vector.size();
    *input_ids = (int*) malloc(sizeof(int) * (*num_input_ids));
    memcpy(*input_ids, input_ids_vector.data(), sizeof(int) * (*num_input_ids));

    // token_type_ids_vector
    *num_token_type_ids = token_type_ids_vector.size();
    *token_type_ids = (int*) malloc(sizeof(int) * (*num_token_type_ids));
    memcpy(*token_type_ids, token_type_ids_vector.data(), sizeof(int) * (*num_token_type_ids));

    // attention_mask_vector
    *num_attention_mask = attention_mask_vector.size();
    *attention_mask = (int*) malloc(sizeof(int) * (*num_attention_mask));
    memcpy(*attention_mask, attention_mask_vector.data(), sizeof(int) * (*num_attention_mask));

    // offsets_vector
    *num_offsets = offsets_vector.size();
    *offsets = (int*) malloc(sizeof(int) * (*num_offsets));
    memcpy(*offsets, offsets_vector.data(), sizeof(int) * (*num_offsets));
}

void wordPieceTokenize(EasyTokenizer tokenizer, char* text, char*** tokens, int* num_tokens, int** offsets, int* num_offsets) {
    std::vector<std::string> token_vector;
    std::vector<tokenizer::SizeT> offset_vector;
    ((tokenizer::Tokenizer *) tokenizer)->wordpiece_tokenize_with_offsets(text, token_vector, offset_vector);

    *num_tokens = token_vector.size();
    *tokens = (char**)malloc(sizeof(char*) * (*num_tokens));
    for (int i = 0; i < *num_tokens; i++) {
        (*tokens)[i] = (char*)malloc(token_vector[i].size() + 1);
        strcpy((*tokens)[i], token_vector[i].c_str());
    }

    *num_offsets = offset_vector.size();
    *offsets = (int*)malloc(sizeof(int) * (*num_offsets));
    memcpy(*offsets, offset_vector.data(), sizeof(int) * (*num_offsets));
}
