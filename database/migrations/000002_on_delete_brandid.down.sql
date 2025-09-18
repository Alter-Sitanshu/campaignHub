ALTER TABLE campaigns
DROP CONSTRAINT fk_campaign_brand;

ALTER TABLE campaigns
ADD CONSTRAINT fk_campaign_brand
FOREIGN KEY (brand_id)
REFERENCES brands (id);