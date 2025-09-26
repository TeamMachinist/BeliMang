-- Insert dummy users
INSERT INTO users (id, username, password_hash, email, role)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'admin01', 'hashedpassword123', 'admin01@example.com', 'admin'),
  ('22222222-2222-2222-2222-222222222222', 'user01', 'hashedpassword456', 'user01@example.com', 'user'),
  ('33333333-3333-3333-3333-333333333333', 'user02', 'hashedpassword789', 'user02@example.com', 'user');

-- Insert dummy merchants
INSERT INTO merchants (id, admin_id, name, merchant_category, image_url, lat, lng)
VALUES
  ('aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1', '11111111-1111-1111-1111-111111111111',
   'Sate Enak', 'SmallRestaurant', 'https://picsum.photos/200/300?random=1.jpg', -6.2000, 106.8167),

  ('aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2', '11111111-1111-1111-1111-111111111111',
   'Bakso Mantap', 'MediumRestaurant', 'https://picsum.photos/200/300?random=2.jpg', -6.2100, 106.8200),

  ('aaaaaaa3-aaaa-aaaa-aaaa-aaaaaaaaaaa3', '11111111-1111-1111-1111-111111111111',
   'Ayam Geprek', 'LargeRestaurant', 'https://picsum.photos/200/300?random=3.jpg', -6.2200, 106.8300),

  ('aaaaaaa4-aaaa-aaaa-aaaa-aaaaaaaaaaa4', '11111111-1111-1111-1111-111111111111',
   'Martabak Boss', 'BoothKiosk', 'https://picsum.photos/200/300?random=4.jpg', -6.2300, 106.8400),

  ('aaaaaaa5-aaaa-aaaa-aaaa-aaaaaaaaaaa5', '11111111-1111-1111-1111-111111111111',
   'Indo Mini', 'ConvenienceStore', 'https://picsum.photos/200/300?random=5.jpg', -6.2400, 106.8500);

