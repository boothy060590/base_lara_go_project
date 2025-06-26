import * as yup from "yup";

export const registerSchema = yup.object().shape({
  password: yup
    .string()
    .required("Password is required!")
    .min(8, "Password must be at least 8 characters.")
    .max(64, "Password must be at most 64 characters."),
  password_confirmation: yup
    .string()
    .required("Password confirmation is required!")
    .oneOf([yup.ref("password")], "Passwords must match!"),
  first_name: yup
    .string()
    .required("First name is required!")
    .matches(
      /^[A-Za-z-' ]+$/,
      "First name must only contain letters, hyphens, apostrophes, and spaces."
    ),
  last_name: yup
    .string()
    .required("Last name is required!")
    .matches(
      /^[A-Za-z-' ]+$/,
      "Last name must only contain letters, hyphens, apostrophes, and spaces."
    ),
  email: yup
    .string()
    .email("Invalid email")
    .required("Email is required!"),
}); 