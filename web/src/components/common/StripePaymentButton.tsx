import { PaymentEventSessionCreatedData } from "@/types/contracts";
import { Button } from "@/components/ui/button";
import { loadStripe } from "@stripe/stripe-js";

interface StripePaymentButtonProps {
  paymentSession: PaymentEventSessionCreatedData;
  isLoading?: boolean;
}

// Initialize Stripe
const stripePromise = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
  ? loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY)
  : null;

export const StripePaymentButton = ({
  paymentSession,
  isLoading = false,
}: StripePaymentButtonProps) => {
  const handlePayment = async () => {
    if (!stripePromise) {
      console.error("Stripe Failed to load: Key not set");
      return;
    }

    const stripe = await stripePromise;
    if (!stripe) {
      console.error("Stripe Failed to load");
      return;
    }

    // Redirect to Stripe Checkout
    const { error } = await stripe.redirectToCheckout({
      sessionId: paymentSession.sessionID,
    });

    if (error) {
      console.error("Payment error:", error);
    }
  };

  if (!process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY) {
    return (
      <Button disabled className="w-full bg-red-500 text-white">
        Stripe API KEY is not set on the NEXTJS app
      </Button>
    );
  }

  return (
    <Button onClick={handlePayment} disabled={isLoading} className="w-full">
      {isLoading
        ? "Loading..."
        : `Pay ${paymentSession.amount} ${paymentSession.currency}`}
    </Button>
  );
};
