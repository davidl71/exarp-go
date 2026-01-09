package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Comment represents a task comment
type Comment struct {
	ID           string
	TaskID       string
	Type         string
	Content      string
	Created      string
	LastModified string
}

// AddComments adds one or more comments to a task
// Uses a transaction to atomically insert all comments
func AddComments(taskID string, comments []Comment) error {
	if len(comments) == 0 {
		return nil // Nothing to do
	}

	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().Format(time.RFC3339)

	for i := range comments {
		comment := &comments[i]

		// Ensure TaskID is set
		if comment.TaskID == "" {
			comment.TaskID = taskID
		}

		// Generate ID if not provided
		if comment.ID == "" {
			timestamp := time.Now().UnixNano()
			comment.ID = fmt.Sprintf("%s-C-%d", taskID, timestamp)
		}

		// Set timestamps if not provided
		if comment.Created == "" {
			comment.Created = now
		}
		if comment.LastModified == "" {
			comment.LastModified = now
		}

		// Insert comment
		_, err = tx.Exec(`
			INSERT INTO task_comments (
				id, task_id, comment_type, content, created, last_modified, created_at
			) VALUES (?, ?, ?, ?, ?, ?, strftime('%s', 'now'))
		`,
			comment.ID,
			comment.TaskID,
			comment.Type,
			comment.Content,
			comment.Created,
			comment.LastModified,
		)
		if err != nil {
			return fmt.Errorf("failed to insert comment %s: %w", comment.ID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetComments retrieves all comments for a task
func GetComments(taskID string) ([]Comment, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	rows, err := db.Query(`
		SELECT id, task_id, comment_type, content, created, last_modified
		FROM task_comments
		WHERE task_id = ?
		ORDER BY created ASC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var lastModified sql.NullString

		if err := rows.Scan(
			&comment.ID,
			&comment.TaskID,
			&comment.Type,
			&comment.Content,
			&comment.Created,
			&lastModified,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		if lastModified.Valid {
			comment.LastModified = lastModified.String
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return comments, nil
}

// GetCommentsByType retrieves all comments of a specific type across all tasks
func GetCommentsByType(commentType string) ([]Comment, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	rows, err := db.Query(`
		SELECT id, task_id, comment_type, content, created, last_modified
		FROM task_comments
		WHERE comment_type = ?
		ORDER BY created ASC
	`, commentType)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments by type: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var lastModified sql.NullString

		if err := rows.Scan(
			&comment.ID,
			&comment.TaskID,
			&comment.Type,
			&comment.Content,
			&comment.Created,
			&lastModified,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		if lastModified.Valid {
			comment.LastModified = lastModified.String
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return comments, nil
}

// DeleteComment deletes a comment by ID
func DeleteComment(commentID string) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	result, err := db.Exec(`DELETE FROM task_comments WHERE id = ?`, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("comment %s not found", commentID)
	}

	return nil
}

// GetCommentsWithTypeFilter retrieves comments for a task with optional type filter
func GetCommentsWithTypeFilter(taskID string, commentType string) ([]Comment, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	rows, err := db.Query(`
		SELECT id, task_id, comment_type, content, created, last_modified
		FROM task_comments
		WHERE task_id = ? AND comment_type = ?
		ORDER BY created ASC
	`, taskID, commentType)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var lastModified sql.NullString

		if err := rows.Scan(
			&comment.ID,
			&comment.TaskID,
			&comment.Type,
			&comment.Content,
			&comment.Created,
			&lastModified,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		if lastModified.Valid {
			comment.LastModified = lastModified.String
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return comments, nil
}

