import { Button } from "@mui/material";
import { Lock, ShieldCheck, UserPlus } from "lucide-react";

type SignInGateProps = {
  onSignIn: () => void;
};

/**
 * Shown in place of the account area when nobody is signed in. It never
 * redirects — it simply explains what the account holds and opens the shared
 * sign-in dialog so the citizen lands right back here afterwards.
 */
export function SignInGate({ onSignIn }: SignInGateProps) {
  return (
    <div className="account-gate surface">
      <span className="account-gate__icon" aria-hidden="true">
        <Lock size={26} strokeWidth={2.2} />
      </span>
      <h2 className="account-gate__title">Sign in to access your account</h2>
      <p className="account-gate__lead">
        Your dashboard, report history, notifications and settings all live here.
        Sign in to pick up where you left off — the rest of NADAA stays open to
        everyone.
      </p>
      <Button
        type="button"
        variant="contained"
        color="warning"
        onClick={onSignIn}
        startIcon={<UserPlus size={18} />}
        className="account-gate__cta signin-submit"
      >
        Sign in to continue
      </Button>
      <p className="account-gate__note">
        <ShieldCheck size={15} aria-hidden="true" />
        Life-safety features — risk, alerts, shelters and 112 — never need an
        account.
      </p>
    </div>
  );
}

export default SignInGate;
