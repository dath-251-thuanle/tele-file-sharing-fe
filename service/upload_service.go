package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	// "time"
	"fe-file-sharing/internal/api"
)

type UploadService struct {
	apiClient *api.Client
}

func NewUploadService(client *api.Client) *UploadService {
	return &UploadService{apiClient: client}
}

// Bước 1: StartUpload - Yêu cầu backend tạo upload và lấy URL
func (s *UploadService) StartUpload(telegramID int64, filename string, size int64) (int64, string, error) {
	// Prepare request to backend using existing api client method UploadFile
	req := api.UploadFileRequest{
		TelegramFileID: "",
		Filename:       filename,
		Size:           size,
	}

	// set telegram id on client temporarily so headers are sent
	prevID := s.apiClient.TelegramID
	s.apiClient.TelegramID = telegramID
	resp, err := s.apiClient.UploadFile(req)
	// restore
	s.apiClient.TelegramID = prevID

	if err != nil {
		return 0, "", fmt.Errorf("không thể khởi tạo upload: %w", err)
	}

	return resp.FileID, resp.UploadURL, nil







	// // --- Code MOCK (Chạy Demo) ---
	// fmt.Println("⚠️ [MOCK] Giả lập StartUpload thành công")
	// fakeFileID := int64(time.Now().Unix()) // Tạo ID ngẫu nhiên theo thời gian
	// fakeUrl := "https://httpbin.org/put"   // Link test upload giả (an toàn)
	
	// return fakeFileID, fakeUrl, nil
}

// Bước 2: UploadToMinIO - Đẩy file Binary trực tiếp lên URL
// Hàm này KHÔNG cần telegramID vì nó upload thẳng lên MinIO/S3 qua URL đã ký
func (s *UploadService) UploadToMinIO(ctx context.Context, presignedURL string, filePath string) error {
	// Đọc file từ ổ đĩa
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("lỗi đọc file tạm: %w", err)
	}

	// Tạo PUT request tới MinIO
	req, err := http.NewRequestWithContext(ctx, "PUT", presignedURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	// MinIO/S3 yêu cầu Content-Type stream cho presigned PUT
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("lỗi mạng khi upload MinIO: %w", err)
	}
	defer resp.Body.Close()

	// Chấp nhận 200 OK hoặc 204 No Content
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("MinIO từ chối nhận file, status: %d", resp.StatusCode)
	}

	return nil






	// // Không upload thật để đỡ tốn thời gian
	// fmt.Println("⚠️ [MOCK] Giả lập UploadToMinIO thành công")
	// // Giả vờ đợi 1 giây cho giống thật
	// time.Sleep(1 * time.Second)
	// return nil
}

// Bước 3: CompleteUpload - Báo Backend biết đã xong
// Cập nhật: Thêm tham số telegramID int64
func (s *UploadService) CompleteUpload(telegramID int64, fileID int64) (*api.File, error) {
	// Call ReportComplete to notify backend
	req := api.ReportUploadCompleteRequest{
		Status: "completed",
	}

	prevID := s.apiClient.TelegramID
	s.apiClient.TelegramID = telegramID
	_, err := s.apiClient.ReportComplete(fileID, req)
	s.apiClient.TelegramID = prevID

	if err != nil {
		return nil, fmt.Errorf("lỗi xác nhận hoàn tất: %w", err)
	}

	// Try to fetch upload report (optional)
	_, err = s.apiClient.GetFileReport(fileID)
	if err != nil {
		// Not fatal: backend may not implement report endpoint
		return nil, nil
	}

	return nil, nil
}







// // CompleteUpload: GIẢ LẬP BÁO CÁO XONG
// func (s *UploadService) CompleteUpload(telegramID int64, fileID int64) error {
// 	fmt.Println("⚠️ [MOCK] Giả lập CompleteUpload thành công")
// 	return nil
// }

// // ListFiles: TRẢ VỀ DANH SÁCH FILE GIẢ ĐỂ TEST NÚT BẤM
// func (s *UploadService) ListFiles(telegramID int64) ([]api.File, error) {
// 	// --- Code thật (Tạm đóng) ---
// 	/*
// 	client := *s.apiClient
// 	client.TelegramID = telegramID
// 	return client.ListFiles()
// 	*/

// 	// --- Code MOCK ---
// 	fmt.Println("⚠️ [MOCK] Trả về danh sách file giả")
// 	mockFiles := []api.File{
// 		{ID: 101, Filename: "Bao_cao_M2.pdf", Size: 2048000},
// 		{ID: 102, Filename: "Anh_ky_yeu.jpg", Size: 512000},
// 		{ID: 103, Filename: "Slide_demo.pptx", Size: 102400},
// 	}
// 	return mockFiles, nil
// }