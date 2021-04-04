import React, { FunctionComponent } from "react";

const Badge: FunctionComponent<{ color?: string }> = ({
  children,
  color = "secondary",
}) => {
  return <span className={`badge badge--${color}`}>{children}</span>;
};

export const RequiredBadge: FunctionComponent = () => (
  <Badge color="danger">Required</Badge>
);

export default Badge;
