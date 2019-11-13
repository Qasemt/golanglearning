-- SELECT COUNT(*) cnt FROM stock_from_web_services 
SELECT Max(time_frame) as ff ,count(*) cnt FROM stock_from_web_services group by time_frame