import os
import pandas as pd
from datetime import datetime
import joblib
import pickle
import numpy as np
from data_cleaning import preprocessing
from scrape import scrape_website
from extract import extract_body_content, clean_body_content
from keras.preprocessing.sequence import pad_sequences
import tensorflow as tf
import json

LOG_FILE = "ml_analyzer_log.json"

# ==========================================
# Logging
# ==========================================
def log_event(domain, status):
    logs = []
    if os.path.exists(LOG_FILE):
        with open(LOG_FILE, "r") as f:
            logs = json.load(f)
    logs.append({
        "domain": domain,
        "status": status,
        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    })
    with open(LOG_FILE, "w") as f:
        json.dump(logs, f, indent=4)
        
# ==========================================
# Load pretrained ML models
# ==========================================
def load_models():
    base_path = "./PreTrained-20251012T082653Z-1-001/PreTrained"

    tfidf_vectorizer = joblib.load(os.path.join(base_path, "tfidf_vectorizer.joblib"))
    nb_model = joblib.load(os.path.join(base_path, "naive_bayes_model.joblib"))
    lstm_model = tf.keras.models.load_model(os.path.join(base_path, "LSTM_pretrained_model.h5"))
    with open(os.path.join(base_path, "tokenizer.pickle"), "rb") as handle:
        tokenizer = pickle.load(handle)

    return tfidf_vectorizer, nb_model, lstm_model, tokenizer

tfidf_vectorizer, nb_model, lstm_model, tokenizer = load_models()

# ==========================================
# ML Analysis
# ==========================================
def analyze_domain(domain):
    try:
        html = scrape_website(domain)
        body = extract_body_content(html)
        clean = clean_body_content(body)

        print(f"[DEBUG SCRAPE] domain={domain}")
        print(f"  html_len={len(html) if html else 0}")
        print(f"  body_len={len(body) if body else 0}")
        print(f"  clean_len={len(clean) if clean else 0}")
        print(f"  clean_sample={clean[:100] if clean else 'EMPTY'}")

        df = pd.DataFrame([{"Website": domain, "Cleaned Content": clean}])
        df_prep = preprocessing(df)
        text_data = df_prep.iloc[:, -1].astype(str).tolist()

        # Validate cleaned content
        if not text_data or not text_data[0].strip():
            print(f"[ML WARN] Empty content for {domain}, marking as Unknown")
            return "Unknown", "Forward"

        # Transform text for Naive Bayes
        # X_tfidf = tfidf_vectorizer.transform(text_data)

        # Tokenize and pad for LSTM
        sequences = tokenizer.texts_to_sequences(text_data)
        padded = pad_sequences(sequences,maxlen=22,padding='post',truncating='post')

        # Naive Bayes: use predict_proba and get positive-class probability
        # prob_nb = 0.0
        # try:
        #     if hasattr(nb_model, "predict_proba"):
        #         probs = nb_model.predict_proba(X_tfidf)[0]
        #         # determine index of positive class (assumes positive label is 1)
        #         if hasattr(nb_model, "classes_"):
        #             classes = list(nb_model.classes_)
        #             pos_idx = classes.index(1) if 1 in classes else (1 if len(classes) > 1 else 0)
        #         else:
        #             pos_idx = 1 if len(probs) > 1 else 0
        #         prob_nb = float(probs[pos_idx])
        #     else:
        #         y_pred_nb = nb_model.predict(X_tfidf)[0]
        #         prob_nb = 1.0 if y_pred_nb == 1 else 0.0
        # except Exception as e:
        #     print(f"[ML WARN] NB probability error: {e}")
        #     prob_nb = 0.0

        # LSTM: probability output (assume single-sigmoid output)
        try:
            pred_lstm = float(lstm_model.predict(padded, verbose=0)[0][0])
        except Exception as e:
            print(f"[ML WARN] LSTM predict error: {e}")
            pred_lstm = 0.0

        # Decide with sensible thresholds (tune these)
        # is_nb = prob_nb >= 0.8
        # is_lstm = pred_lstm >= 0.7
        # result = 1 if (is_lstm) else 0
        if pred_lstm >= 0.7:
            result = 1
        else:
            result = 0

        # Debug info to help diagnose why everything is flagged
        print(f"[ML DEBUG] domain={domain} clean_len={len(text_data[0])} prob_nb={0.0:.3f} pred_lstm={pred_lstm:.3f} result={result}")
        if result == 1:
            print(f"[ML DETECT] {domain} flagged as gambling ({pred_lstm*100:.2f}%)")
            return "Blacklist", "Block"
        else:
            print(f"[ML DETECT] {domain} flagged as non-gambling ({pred_lstm*100:.2f}%)")
            return "Whitelist", "Allow"


    except Exception as e:
        print(f"[ML ERROR] {domain}: {e}")
        return "Unknown", "Forward"


