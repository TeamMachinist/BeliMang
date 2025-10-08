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

-- Items for Sate Enak (merchant aaaaaaa1-...)
INSERT INTO items (id, merchant_id, name, product_category, price, image_url)
VALUES
  ('bbbbbbb1-1111-1111-1111-bbbbbbbbbbb1', 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1',
   'Sate Ayam', 'Makanan', 25000, 'https://picsum.photos/200/200?random=sate1.jpg'),
  ('bbbbbbb1-1111-1111-1111-bbbbbbbbbbb2', 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1',
   'Sate Kambing', 'Makanan', 35000, 'https://picsum.photos/200/200?random=sate2.jpg');

-- Items for Bakso Mantap (merchant aaaaaaa2-...)
INSERT INTO items (id, merchant_id, name, product_category, price, image_url)
VALUES
  ('bbbbbbb2-2222-2222-2222-bbbbbbbbbbb1', 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2',
   'Bakso Urat', 'Makanan', 20000, 'https://picsum.photos/200/200?random=bakso1.jpg'),
  ('bbbbbbb2-2222-2222-2222-bbbbbbbbbbb2', 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2',
   'Bakso Telur', 'Makanan', 25000, 'https://picsum.photos/200/200?random=bakso2.jpg');

-- Items for Ayam Geprek (merchant aaaaaaa3-...)
INSERT INTO items (id, merchant_id, name, product_category, price, image_url)
VALUES
  ('bbbbbbb3-3333-3333-3333-bbbbbbbbbbb1', 'aaaaaaa3-aaaa-aaaa-aaaa-aaaaaaaaaaa3',
   'Ayam Geprek Original', 'Makanan', 18000, 'https://picsum.photos/200/200?random=geprek1.jpg'),
  ('bbbbbbb3-3333-3333-3333-bbbbbbbbbbb2', 'aaaaaaa3-aaaa-aaaa-aaaa-aaaaaaaaaaa3',
   'Ayam Geprek Level 10', 'Makanan', 22000, 'https://picsum.photos/200/200?random=geprek2.jpg');

-- Items for Martabak Boss (merchant aaaaaaa4-...)
INSERT INTO items (id, merchant_id, name, product_category, price, image_url)
VALUES
  ('bbbbbbb4-4444-4444-4444-bbbbbbbbbbb1', 'aaaaaaa4-aaaa-aaaa-aaaa-aaaaaaaaaaa4',
   'Martabak Manis', 'Makanan', 30000, 'https://picsum.photos/200/200?random=martabak1.jpg'),
  ('bbbbbbb4-4444-4444-4444-bbbbbbbbbbb2', 'aaaaaaa4-aaaa-aaaa-aaaa-aaaaaaaaaaa4',
   'Martabak Telur', 'Makanan', 35000, 'https://picsum.photos/200/200?random=martabak2.jpg');

-- Items for Indo Mini (merchant aaaaaaa5-...)
INSERT INTO items (id, merchant_id, name, product_category, price, image_url)
VALUES
  ('bbbbbbb5-5555-5555-5555-bbbbbbbbbbb1', 'aaaaaaa5-aaaa-aaaa-aaaa-aaaaaaaaaaa5',
   'Air Mineral 600ml', 'Minuman', 5000, 'https://picsum.photos/200/200?random=air.jpg'),
  ('bbbbbbb5-5555-5555-5555-bbbbbbbbbbb2', 'aaaaaaa5-aaaa-aaaa-aaaa-aaaaaaaaaaa5',
   'Chiki Balls', 'Snack', 3000, 'https://picsum.photos/200/200?random=chiki.jpg');
