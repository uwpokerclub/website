import type { SemesterSetupFormData } from "./schema";
import styles from "./SemesterSetupWizard.module.css";

interface ReviewStepProps {
  data: SemesterSetupFormData;
  name: string;
}

function ReviewValue({ label, value, dataQa }: { label: string; value: string | number; dataQa: string }) {
  return (
    <div className={styles.reviewItem}>
      <dt>{label}</dt>
      <dd data-qa={dataQa}>{value}</dd>
    </div>
  );
}

export function ReviewStep({ data, name }: ReviewStepProps) {
  const term = `${data.term.charAt(0).toUpperCase()}${data.term.slice(1)}`;

  return (
    <section data-qa="semester-wizard-step-review">
      <p className={styles.stepDescription}>Review the semester before creating it.</p>
      <dl className={styles.reviewList}>
        <ReviewValue label="Semester Name" value={name} dataQa="semester-review-name" />
        <ReviewValue label="Term" value={term} dataQa="semester-review-term" />
        <ReviewValue label="Start Date" value={data.startDate} dataQa="semester-review-startDate" />
        <ReviewValue label="End Date" value={data.endDate} dataQa="semester-review-endDate" />
        <ReviewValue
          label="Starting Budget"
          value={`$${data.startingBudget}`}
          dataQa="semester-review-startingBudget"
        />
        <ReviewValue label="Membership Fee" value={`$${data.membershipFee}`} dataQa="semester-review-membershipFee" />
        <ReviewValue
          label="Discounted Membership Fee"
          value={`$${data.membershipDiscountFee}`}
          dataQa="semester-review-membershipDiscountFee"
        />
        <ReviewValue label="Rebuy Fee" value={`$${data.rebuyFee}`} dataQa="semester-review-rebuyFee" />
        <ReviewValue label="Free Trial Events" value={data.freeTrialLimit} dataQa="semester-review-freeTrialLimit" />
        {data.meta && <ReviewValue label="Additional Details" value={data.meta} dataQa="semester-review-meta" />}
      </dl>
    </section>
  );
}
