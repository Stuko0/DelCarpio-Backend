-- ── Payments ──
CREATE TABLE IF NOT EXISTS public.payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES public.orders(id) ON DELETE CASCADE,
    mp_preference_id TEXT,
    mp_payment_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    amount NUMERIC(10,2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'BOB',
    created TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON public.payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_mp_payment_id ON public.payments(mp_payment_id);

ALTER TABLE public.payments ENABLE ROW LEVEL SECURITY;

CREATE POLICY payment_user_policy ON public.payments
    FOR SELECT
    TO authenticated
    USING (
        order_id IN (SELECT id FROM public.orders WHERE user_id = auth.uid())
    );
