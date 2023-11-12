CREATE TABLE companies (
name TEXT PRIMARY KEY,
is_top_500 BOOLEAN,
website_url TEXT,
linkedin_url TEXT,
wttj_url TEXT,
job_page_url TEXT,
last_offers_update DATE
);

CREATE TABLE offers (
id SERIAL PRIMARY KEY,
company_name TEXT,
offer_url TEXT,
UNIQUE(offer_url)
);