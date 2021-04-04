import React, { FunctionComponent, ReactNode } from "react";
import clsx from "clsx";
import Layout from "@theme/Layout";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import { usePluginData } from "@docusaurus/useGlobalData";
import useBaseUrl from "@docusaurus/useBaseUrl";
import styles from "./styles.module.scss";

interface FeatureProps {
  title: string;
  imageUrl: string;
  description: ReactNode;
}

const features: FeatureProps[] = [
  {
    title: "Trigger by Webhooks",
    imageUrl: "img/undraw_data_processing_yrrv.svg",
    description: (
      <>Create, Update or delete Kubernetes resources by HTTP webhooks.</>
    ),
  },
  {
    title: "Copy Resources",
    imageUrl: "img/undraw_Documents_re_isxv.svg",
    description: (
      <>
        Copy existing Kubernetes resources and mutate them with{" "}
        <a
          href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md"
          target="_blank"
          rel="noreferrer"
        >
          Strategic Merge Patch
        </a>{" "}
        or{" "}
        <a href="http://jsonpatch.com/" target="_blank" rel="noreferrer">
          JSON Patch
        </a>
        .
      </>
    ),
  },
  {
    title: "GitHub Integration",
    imageUrl: "img/undraw_version_control_9bpv.svg",
    description: <>GitHub push and pull request events are also supported.</>,
  },
];

const Feature: FunctionComponent<FeatureProps> = ({
  imageUrl,
  title,
  description,
}) => {
  const imgUrl = useBaseUrl(imageUrl);

  return (
    <div className={clsx("col col--4", styles.feature)}>
      {imgUrl && (
        <div className="text--center">
          <img className={styles.featureImage} src={imgUrl} alt={title} />
        </div>
      )}
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  );
};

const Home: FunctionComponent = () => {
  const context = useDocusaurusContext();
  const docConfig = usePluginData("docusaurus-plugin-content-docs");
  const { siteConfig = {} } = context;
  const currentVersion = docConfig.versions.find(
    (ver) => ver.name === "current"
  );
  const entryDoc = currentVersion.docs.find(
    (doc) => doc.id === currentVersion.mainDocId
  );

  return (
    <Layout>
      <header className={clsx("hero", styles.heroBanner)}>
        <div className="container">
          <h1 className="hero__title">{siteConfig.title}</h1>
          <p className="hero__subtitle">{siteConfig.tagline}</p>
          <div className={styles.buttons}>
            <Link
              className={clsx(
                "button button--outline button--primary button--lg",
                styles.getStarted
              )}
              to={entryDoc.path}
            >
              Get Started
            </Link>
          </div>
        </div>
      </header>
      <main>
        {features && features.length > 0 && (
          <section className={styles.features}>
            <div className="container">
              <div className="row">
                {features.map((props, idx) => (
                  <Feature key={idx} {...props} />
                ))}
              </div>
            </div>
          </section>
        )}
      </main>
    </Layout>
  );
};

export default Home;
