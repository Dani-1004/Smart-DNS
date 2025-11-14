import nltk
import pandas as pd
from nltk.tokenize import word_tokenize
from nltk.corpus import stopwords
from Sastrawi.Stemmer.StemmerFactory import StemmerFactory
import swifter


def preprocessing(df):
    
    nltk.download('punkt', quiet=True)
    nltk.download('punkt_tab', quiet=True)
    nltk.download('stopwords', quiet=True)

    # Bersihkan data
    df_clean_false = df[~df['Cleaned Content'].str.contains('NA', na=False)].dropna()
    df_clean_false['Cleansing'] = df_clean_false['Cleaned Content'].str.replace(r'\n', ' ', regex=True)

    # Label positif
    positive_keywords = ['slot', 'bonus', 'win', 'gacor', 'maxwin']
    df_clean_false['Label'] = df_clean_false['Cleaned Content'].apply(
        lambda x: 'Positive' if any(keyword in x.lower() for keyword in positive_keywords) else 'Negative'
    )

    # Case folding
    df_clean_false['casefolding'] = df_clean_false['Cleansing'].str.lower()

    # Tokenizing
    df_clean_false['tokenize'] = df_clean_false['casefolding'].apply(word_tokenize)

    # Normalisasi kata
    dataset2 = './Data_Processing/KamusKata.csv'
    dict_word = pd.read_csv(dataset2, encoding='latin-1', header=None)
    dict_word = dict_word.rename(columns={0: 'original', 1: 'replacement'})

    normalized_word_dict = {}
    for _, row in dict_word.iterrows():
        normalized_word_dict[row.iloc[0]] = row.iloc[1]

    def normalized_term(document):
        return [normalized_word_dict.get(term, term) for term in document]

    df_clean_false['normalisasi'] = df_clean_false['tokenize'].apply(normalized_term)

    # Stopwords
    list_stopwords = stopwords.words('indonesian')
    list_stopwords.extend([
        "yg","dg","rt","dgn","ny","d","klo","kalo","amp","biar","bikin","bilang",
        "gak","ga","krn","nya","nih","sih","si","tau","tdk","tuh","utk","ya",
        "jd","jgn","sdh","aja","n","t","nyg","hehe","pen","u","nan","loh","rt",
        "&amp","yah","gtgt","ltlt"
    ])

    # Baca stopword tambahan
    txt_stopword = pd.read_csv('./Data_Processing/stopwords.txt', sep=" ", header=None, encoding='utf-8')
    if not txt_stopword.empty:
        txt_words = txt_stopword.iloc[0, 0].split(' ')
        list_stopwords.extend(txt_words)

    list_stopwords = set(list_stopwords)

    def stopwords_removal(words):
        return [word for word in words if word not in list_stopwords]

    df_clean_false['stopwords'] = df_clean_false['normalisasi'].apply(stopwords_removal)

    # Stemming
    factory = StemmerFactory()
    stemmer = factory.create_stemmer()

    def stemmed_wrapper(term):
        return stemmer.stem(term)

    term_dict = {}
    for document in df_clean_false['stopwords']:
        for term in document:
            if term not in term_dict:
                term_dict[term] = stemmed_wrapper(term)

    def get_stemmed_term(document):
        return [term_dict[term] for term in document]

    df_clean_false['stemming'] = df_clean_false['stopwords'].swifter.apply(get_stemmed_term)

    return df_clean_false
