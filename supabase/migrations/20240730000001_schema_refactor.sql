-- ── Fase 1.1: Rename profiles → users, add role + metadata columns ──
ALTER TABLE IF EXISTS public.profiles RENAME TO users;
UPDATE public.users SET name = '' WHERE name IS NULL;
UPDATE public.users SET email = '' WHERE email IS NULL;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'customer';
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS avatar_url TEXT;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS phone_verified BOOLEAN DEFAULT false;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS last_login TIMESTAMPTZ;
ALTER TABLE public.users ALTER COLUMN name SET NOT NULL;
ALTER TABLE public.users ALTER COLUMN email SET NOT NULL;

-- ── Fase 1.2: Add product columns ──
ALTER TABLE public.products ADD COLUMN IF NOT EXISTS price NUMERIC(10,2) NOT NULL DEFAULT 0;
ALTER TABLE public.products ADD COLUMN IF NOT EXISTS compare_at_price NUMERIC(10,2);
ALTER TABLE public.products ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE public.products ADD COLUMN IF NOT EXISTS image TEXT;
ALTER TABLE public.products ADD COLUMN IF NOT EXISTS category TEXT;
ALTER TABLE public.products ADD COLUMN IF NOT EXISTS tags TEXT[];
ALTER TABLE public.products ADD COLUMN IF NOT EXISTS stock INTEGER NOT NULL DEFAULT 0;
ALTER TABLE public.products ADD COLUMN IF NOT EXISTS sku TEXT UNIQUE;
ALTER TABLE public.products ADD COLUMN IF NOT EXISTS weight_grams NUMERIC(8,2);

-- ── Fase 1.3: Add recipe columns ──
ALTER TABLE public.recipes ADD COLUMN IF NOT EXISTS image TEXT;
ALTER TABLE public.recipes ADD COLUMN IF NOT EXISTS summary TEXT;
ALTER TABLE public.recipes ADD COLUMN IF NOT EXISTS prep_time TEXT;
ALTER TABLE public.recipes ADD COLUMN IF NOT EXISTS difficulty TEXT;
ALTER TABLE public.recipes ADD COLUMN IF NOT EXISTS servings INTEGER;
ALTER TABLE public.recipes ADD COLUMN IF NOT EXISTS ingredients JSONB NOT NULL DEFAULT '[]';
ALTER TABLE public.recipes ADD COLUMN IF NOT EXISTS instructions JSONB NOT NULL DEFAULT '[]';
ALTER TABLE public.recipes DROP COLUMN IF EXISTS content_markdown;

-- ── Fase 1.4: Add order columns ──
ALTER TABLE public.orders ADD COLUMN IF NOT EXISTS updated TIMESTAMPTZ;
ALTER TABLE public.orders ADD COLUMN IF NOT EXISTS shipping_address JSONB;
ALTER TABLE public.orders ADD COLUMN IF NOT EXISTS payment_method TEXT;
ALTER TABLE public.orders ADD COLUMN IF NOT EXISTS notes TEXT;
ALTER TABLE public.orders ADD COLUMN IF NOT EXISTS order_number TEXT UNIQUE;

-- ── Fase 1.5: Rename payment columns to Stripe ──
ALTER TABLE public.payments RENAME COLUMN mp_preference_id TO stripe_session_id;
ALTER TABLE public.payments RENAME COLUMN mp_payment_id TO stripe_payment_intent_id;
ALTER TABLE public.payments ADD COLUMN IF NOT EXISTS receipt_url TEXT;

-- ── Fase 1.6: Add address columns ──
ALTER TABLE public.addresses ADD COLUMN IF NOT EXISTS country TEXT NOT NULL DEFAULT 'Bolivia';
ALTER TABLE public.addresses ADD COLUMN IF NOT EXISTS zip_code TEXT;
ALTER TABLE public.addresses ADD COLUMN IF NOT EXISTS label TEXT DEFAULT 'Casa';
