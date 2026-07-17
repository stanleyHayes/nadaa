import { CheckCircle2, ChevronRight, ShieldCheck } from "lucide-react";
import { type FormEvent, useState } from "react";
import { marketingLinks } from "@/app/config";

const regions = [
  "Greater Accra",
  "Ashanti",
  "Western",
  "Western North",
  "Central",
  "Eastern",
  "Volta",
  "Oti",
  "Northern",
  "North East",
  "Savannah",
  "Upper East",
  "Upper West",
  "Bono",
  "Bono East",
  "Ahafo",
];

const languages = [
  { value: "en", label: "English" },
  { value: "tw", label: "Twi" },
  { value: "ga", label: "Ga" },
  { value: "ee", label: "Ewe" },
  { value: "dag", label: "Dagbani" },
  { value: "ha", label: "Hausa" },
];

type SignupForm = {
  name: string;
  phone: string;
  region: string;
  language: string;
  consent: boolean;
};

const emptyForm: SignupForm = {
  name: "",
  phone: "",
  region: "",
  language: "en",
  consent: false,
};

function validate(form: SignupForm) {
  const errors: Partial<Record<keyof SignupForm, string>> = {};
  if (form.name.trim().length < 2) {
    errors.name = "Enter your full name.";
  }
  const phone = form.phone.replace(/\s+/g, "");
  if (!/^(\+233|0)\d{9}$/.test(phone)) {
    errors.phone = "Enter a Ghana phone number, e.g. 024 123 4567.";
  }
  if (!form.region) {
    errors.region = "Choose your region.";
  }
  if (!form.consent) {
    errors.consent = "Please acknowledge the preview notice to continue.";
  }
  return errors;
}

export function SignupPage() {
  const [form, setForm] = useState<SignupForm>(emptyForm);
  const [errors, setErrors] = useState<
    Partial<Record<keyof SignupForm, string>>
  >({});
  const [submitted, setSubmitted] = useState(false);

  const update = <K extends keyof SignupForm>(key: K, value: SignupForm[K]) => {
    setForm((current) => ({ ...current, [key]: value }));
    setErrors((current) => ({ ...current, [key]: undefined }));
  };

  const onSubmit = (event: FormEvent) => {
    event.preventDefault();
    const nextErrors = validate(form);
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length === 0) {
      setSubmitted(true);
    }
  };

  if (submitted) {
    return (
      <section className="signup-section" aria-labelledby="signup-done-title">
        <div className="signup-done">
          <span className="signup-done-icon" aria-hidden="true">
            <CheckCircle2 size={30} />
          </span>
          <p className="eyebrow">Sign-up preview</p>
          <h1 id="signup-done-title">
            Welcome, {form.name.trim().split(" ")[0]}.
          </h1>
          <p>
            This preview does not create an account — your details were not
            submitted or stored. When public registration opens, sign up in the
            citizen app to receive flood and fire warnings for {form.region}.
          </p>
          <div className="hero-actions">
            <a className="primary-action" href={marketingLinks.citizenWeb}>
              Open the citizen app
              <ChevronRight aria-hidden="true" size={18} />
            </a>
            <a className="ghost-action" href={marketingLinks.emergencyPhone}>
              Emergency? Call 112
            </a>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section className="signup-section" aria-labelledby="signup-title">
      <div className="signup-copy">
        <p className="eyebrow">Citizen sign-up preview</p>
        <h1 id="signup-title">Get warnings where you live.</h1>
        <p>
          A free citizen account brings flood and fire warnings, area risk
          checks, and incident reporting — in your language, online or offline.
          This page previews the sign-up flow: it does not create an account,
          send messages, or store your details.
        </p>
        <ul className="signup-points">
          <li>
            <ShieldCheck aria-hidden="true" size={18} />
            <span>
              We ask for the minimum needed. Your data is handled under Ghana's
              Data Protection Act, 2012 (Act 843).
            </span>
          </li>
          <li>
            <ShieldCheck aria-hidden="true" size={18} />
            <span>
              You can report anonymously any time, and life-safety alerts are
              never sold or used for advertising.
            </span>
          </li>
        </ul>
      </div>

      <form className="signup-form" noValidate onSubmit={onSubmit}>
        <p className="field-hint">
          Sign-up preview — nothing you enter here is submitted or stored.
        </p>
        <div className="field">
          <label htmlFor="signup-name">Full name</label>
          <input
            aria-describedby={errors.name ? "err-name" : undefined}
            aria-invalid={Boolean(errors.name)}
            autoComplete="name"
            id="signup-name"
            onChange={(event) => update("name", event.target.value)}
            type="text"
            value={form.name}
          />
          {errors.name ? (
            <p className="field-error" id="err-name">
              {errors.name}
            </p>
          ) : null}
        </div>

        <div className="field">
          <label htmlFor="signup-phone">Phone number</label>
          <input
            aria-describedby={errors.phone ? "err-phone" : "hint-phone"}
            aria-invalid={Boolean(errors.phone)}
            autoComplete="tel"
            id="signup-phone"
            inputMode="tel"
            onChange={(event) => update("phone", event.target.value)}
            placeholder="024 123 4567"
            type="tel"
            value={form.phone}
          />
          {errors.phone ? (
            <p className="field-error" id="err-phone">
              {errors.phone}
            </p>
          ) : (
            <p className="field-hint" id="hint-phone">
              Preview only — no verification text is sent from this page.
            </p>
          )}
        </div>

        <div className="field-row">
          <div className="field">
            <label htmlFor="signup-region">Region</label>
            <select
              aria-describedby={errors.region ? "err-region" : undefined}
              aria-invalid={Boolean(errors.region)}
              id="signup-region"
              onChange={(event) => update("region", event.target.value)}
              value={form.region}
            >
              <option value="">Select a region</option>
              {regions.map((region) => (
                <option key={region} value={region}>
                  {region}
                </option>
              ))}
            </select>
            {errors.region ? (
              <p className="field-error" id="err-region">
                {errors.region}
              </p>
            ) : null}
          </div>

          <div className="field">
            <label htmlFor="signup-language">Preferred language</label>
            <select
              id="signup-language"
              onChange={(event) => update("language", event.target.value)}
              value={form.language}
            >
              {languages.map((language) => (
                <option key={language.value} value={language.value}>
                  {language.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="field field-consent">
          <label htmlFor="signup-consent">
            <input
              aria-describedby={errors.consent ? "err-consent" : undefined}
              aria-invalid={Boolean(errors.consent)}
              checked={form.consent}
              id="signup-consent"
              onChange={(event) => update("consent", event.target.checked)}
              type="checkbox"
            />
            <span>
              I understand this preview does not create an account and that
              nothing I enter is submitted or stored.
            </span>
          </label>
          {errors.consent ? (
            <p className="field-error" id="err-consent">
              {errors.consent}
            </p>
          ) : null}
        </div>

        <button className="primary-action signup-submit" type="submit">
          Preview sign-up
          <ChevronRight aria-hidden="true" size={18} />
        </button>
        <p className="field-hint">
          In a life-threatening emergency, call 112 — do not wait for sign-up.
        </p>
      </form>
    </section>
  );
}
