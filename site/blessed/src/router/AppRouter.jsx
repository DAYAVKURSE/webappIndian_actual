// src/app/router/AppRouter.jsx
import { Routes, Route } from 'react-router-dom';
import { appRoutes } from './routes';

export const AppRouter = () => {
  const renderRoutes = (routes) =>
    routes.map(({ path, element, index, children }) => (
      <Route
        key={path || Math.random()}
        path={path}
        element={element}
        index={index}
      >
        {children && renderRoutes(children)}
      </Route>
    ));

  return <Routes>{renderRoutes(appRoutes)}</Routes>;
};
