cd frontend
echo "Building Frontend"
npm run build
cd ..
echo "Copy the static files"
mkdir -p static
cp -r frontend/public/* static/
echo "Finished"