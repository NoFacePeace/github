import cv2

# 读取图像
image = cv2.imread('image.jpg')
gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)

# 应用 Canny 边缘检测
edges = cv2.Canny(gray, 50, 150)

# 找到轮廓
contours, _ = cv2.findContours(edges, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)

# 绘制轮廓
for contour in contours:
    cv2.drawContours(image, [contour], -1, (0, 255, 0), 2)

# 显示结果
cv2.imshow('Contours', image)
cv2.waitKey(0)
cv2.destroyAllWindows()
