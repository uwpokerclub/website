import { Component, ReactNode } from "react";
import { Button } from "@uwpokerclub/components";
import gooseError from "@/assets/goose-error.webp";
import styles from "./ErrorBoundary.module.css";

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  reset = () => this.setState({ error: null });

  render() {
    const { error } = this.state;
    const { children, fallback } = this.props;

    if (!error) return children;

    if (fallback) return fallback;

    return (
      <div className={styles.container}>
        <img src={gooseError} alt="" className={styles.illustration} />
        <h2 className={styles.title}>Honk! You called our bluff.</h2>
        <p className={styles.subtitle}>Something went wrong. If the issue persists, please contact the Webmaster.</p>
        <p className={styles.message}>{error.message}</p>
        <Button onClick={this.reset}>Try again</Button>
      </div>
    );
  }
}
