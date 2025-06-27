# Frontend Application

A modern Vue.js 3 frontend application with reactive forms, authentication, and integration with the async Go backend.

## 🚀 Features

- **Vue.js 3**: Modern reactive frontend framework
- **Vite**: Fast build tool and development server
- **SCSS**: Modern CSS with variables and mixins
- **Form Validation**: Client-side validation with custom validators
- **Authentication**: JWT-based login/register system
- **Responsive Design**: Mobile-first responsive layout
- **Hot Module Replacement**: Instant development feedback

## 🏗️ Architecture

### Component Structure

```
src/
├── components/          # Reusable UI components
│   └── form/           # Form field components
├── Pages/              # Page components
│   ├── auth/           # Authentication pages
│   └── home/           # Dashboard pages
├── helpers/            # Utility functions
│   └── api/            # API integration
├── store/              # State management
├── form_validators/    # Form validation logic
└── assets/             # Static assets and styles
```

### State Management

- **Pinia**: Modern state management for Vue 3
- **Auth Store**: Manages authentication state
- **Reactive Updates**: Automatic UI updates on state changes

## 🛠️ Technology Stack

- **Vue.js 3**: Core framework
- **Vite**: Build tool and dev server
- **SCSS**: Styling with modern features
- **Axios**: HTTP client for API calls
- **Pinia**: State management
- **Font Awesome**: Icon library

## 🚀 Quick Start

### Prerequisites

- Node.js 20+
- npm or yarn

### Installation

1. **Install dependencies**
   ```bash
   cd frontend
   npm install
   ```

2. **Start development server**
   ```bash
   npm run dev
   ```

3. **Access the application**
   - **Development**: http://localhost:5173
   - **Production**: https://app.baselaragoproject.test

## 📁 Project Structure

### Components

#### Form Components (`src/components/form/`)

- **EmailFormField.vue**: Email input with validation
- **PasswordFormField.vue**: Password input with strength indicator
- **TelephoneFormField.vue**: Phone number input with formatting
- **TextFormField.vue**: Generic text input with validation

#### Page Components (`src/Pages/`)

- **Auth Pages** (`src/Pages/auth/`)
  - **Login.vue**: User login form
  - **Register.vue**: User registration form
- **Home Pages** (`src/Pages/home/`)
  - **Home.vue**: Main dashboard
  - **Admin.vue**: Admin dashboard
  - **Customer.vue**: Customer dashboard

### API Integration (`src/helpers/api/`)

- **api.js**: Base API configuration
- **auth/authApi.js**: Authentication API calls

### Form Validation (`src/form_validators/`)

- **index.js**: Validator utilities
- **login_validator.js**: Login form validation
- **register_validator.js**: Registration form validation

### State Management (`src/store/`)

- **auth.js**: Authentication state management

## 🔧 Configuration

### Environment Variables

```env
# API Configuration
VITE_API_BASE_URL=https://api.baselaragoproject.test
VITE_APP_NAME=Base Laravel Go Project
```

### API Integration

The frontend integrates with the Go backend's async event and queue system:

#### Authentication Flow

1. **User Registration**
   - Frontend sends registration data to API
   - API creates user and dispatches `UserCreated` event
   - Event triggers async email sending
   - User receives welcome email via queue

2. **User Login**
   - Frontend sends login credentials
   - API validates and returns JWT token
   - Frontend stores token in Pinia store
   - Token used for authenticated requests

#### Real-time Updates

- **Async Processing**: Backend processes events and emails asynchronously
- **Queue Monitoring**: Backend handles queue processing automatically
- **Email Delivery**: Welcome emails sent via dedicated mail queue

## 🎨 Styling

### SCSS Structure (`src/assets/scss/`)

- **button.scss**: Button component styles
- **card.scss**: Card component styles
- **form.scss**: Form component styles
- **utilities.scss**: Utility classes

### Design System

- **Color Palette**: Consistent color scheme
- **Typography**: Unified font system
- **Spacing**: Consistent spacing scale
- **Components**: Reusable component library

## 📱 Responsive Design

### Mobile-First Approach

- **Breakpoints**: Mobile, tablet, desktop
- **Flexible Layouts**: CSS Grid and Flexbox
- **Touch-Friendly**: Optimized for mobile interaction

### Component Responsiveness

- **Form Fields**: Adaptive input sizing
- **Navigation**: Collapsible mobile menu
- **Content**: Responsive content layout

## 🔒 Security

### Authentication

- **JWT Tokens**: Secure token-based authentication
- **Token Storage**: Secure token management
- **Route Protection**: Guarded routes for authenticated users

### Form Security

- **Input Validation**: Client-side validation
- **CSRF Protection**: Built-in CSRF protection
- **XSS Prevention**: Sanitized input handling

## 🧪 Testing

### Form Validation Testing

```javascript
// Test email validation
const emailValidator = new EmailValidator();
expect(emailValidator.validate('test@example.com')).toBe(true);
expect(emailValidator.validate('invalid-email')).toBe(false);
```

### API Integration Testing

```javascript
// Test authentication API
const response = await authApi.login({
  email: 'test@example.com',
  password: 'password123'
});
expect(response.token).toBeDefined();
```

## 🚀 Development

### Hot Module Replacement

- **Instant Updates**: Changes reflect immediately
- **State Preservation**: Component state maintained
- **Error Overlay**: Clear error reporting

### Development Tools

- **Vue DevTools**: Component inspection
- **Browser DevTools**: Network and console debugging
- **Vite Dev Server**: Fast development server

### Code Quality

- **ESLint**: Code linting and formatting
- **Prettier**: Code formatting
- **TypeScript**: Type safety (optional)

## 📦 Build and Deployment

### Development Build

```bash
npm run dev
```

### Production Build

```bash
npm run build
```

### Preview Production Build

```bash
npm run preview
```

## 🔄 Integration with Backend

### Event-Driven Architecture

The frontend integrates seamlessly with the backend's event-driven system:

#### User Registration Example

```javascript
// Frontend sends registration data
const response = await authApi.register({
  first_name: 'John',
  last_name: 'Doe',
  email: 'john@example.com',
  password: 'password123',
  password_confirmation: 'password123'
});

// Backend processes asynchronously:
// 1. Creates user in database
// 2. Dispatches UserCreated event
// 3. Event listener queues welcome email
// 4. Mail queue processes and sends email
```

#### Real-time Feedback

- **Immediate Response**: API responds immediately
- **Background Processing**: Events and emails processed asynchronously
- **User Experience**: No waiting for email sending

### API Error Handling

```javascript
try {
  const response = await authApi.register(userData);
  // Handle success
} catch (error) {
  if (error.response?.data?.errors) {
    // Handle validation errors
    setErrors(error.response.data.errors);
  } else {
    // Handle general errors
    setError('Registration failed. Please try again.');
  }
}
```

## 🎯 Best Practices

### Component Design

- **Single Responsibility**: Each component has one purpose
- **Reusability**: Components are reusable and configurable
- **Composition**: Use composition over inheritance

### State Management

- **Centralized State**: Use Pinia for global state
- **Local State**: Use component state for local data
- **Reactive Updates**: Leverage Vue's reactivity system

### Performance

- **Lazy Loading**: Load components on demand
- **Code Splitting**: Split code by routes
- **Optimized Builds**: Use Vite's optimization features

## 🔮 Future Enhancements

### Planned Features

1. **Real-time Notifications**: WebSocket integration
2. **Offline Support**: Service worker implementation
3. **Progressive Web App**: PWA capabilities
4. **Advanced Forms**: Dynamic form generation
5. **Theme System**: Dark/light mode support

### Performance Improvements

1. **Virtual Scrolling**: For large data sets
2. **Image Optimization**: Lazy loading and compression
3. **Bundle Optimization**: Tree shaking and code splitting
4. **Caching Strategy**: Intelligent caching

## 📚 Resources

- [Vue.js 3 Documentation](https://vuejs.org/)
- [Vite Documentation](https://vitejs.dev/)
- [Pinia Documentation](https://pinia.vuejs.org/)
- [SCSS Documentation](https://sass-lang.com/)
